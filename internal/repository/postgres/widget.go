package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

const (
	widgetColumns = `id, header_th, header_en, sub_header_th, sub_header_en,
		start_date, end_date, product_list_ref_id, created_at, created_by, updated_at, updated_by`

	insertWidget = `INSERT INTO widgets (` + widgetColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	updateWidget = `UPDATE widgets SET header_th = $2, header_en = $3, sub_header_th = $4, sub_header_en = $5,
		start_date = $6, end_date = $7, product_list_ref_id = $8, updated_at = $9, updated_by = $10
		WHERE id = $1`

	selectWidget = `SELECT ` + widgetColumns + ` FROM widgets WHERE id = $1`

	selectWidgetPage = `SELECT ` + widgetColumns + ` FROM widgets
		ORDER BY updated_at DESC, id LIMIT $1 OFFSET $2`

	countWidgets = `SELECT count(*) FROM widgets`

	deleteWidget = `DELETE FROM widgets WHERE id = $1`

	insertComponent = `INSERT INTO widget_components
		(widget_id, position, type, sub_type, image_url, is_active, data_src)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	deleteComponents = `DELETE FROM widget_components WHERE widget_id = $1`

	selectComponents = `SELECT widget_id, type, sub_type, image_url, is_active, data_src
		FROM widget_components WHERE widget_id = ANY($1) ORDER BY widget_id, position`
)

// Create writes the row and its components in one transaction.
func (r *Repo) CreateWidget(ctx context.Context, w domain.Widget) error {
	return joinOrBeginTx(ctx, r.pool, func(ctx context.Context) error {
		db := conn(ctx, r.pool)
		if _, err := db.Exec(ctx, insertWidget, w.ID, w.HeaderTH, w.HeaderEN, w.SubHeaderTH, w.SubHeaderEN,
			w.StartDate, w.EndDate, w.ProductListRefID, w.CreatedAt, w.CreatedBy, w.UpdatedAt, w.UpdatedBy); err != nil {
			return mapErr(err)
		}
		return r.insertComponents(ctx, db, w)
	})
}

// Update replaces the editable columns and the component list.
func (r *Repo) UpdateWidget(ctx context.Context, w domain.Widget) error {
	return joinOrBeginTx(ctx, r.pool, func(ctx context.Context) error {
		db := conn(ctx, r.pool)
		tag, err := db.Exec(ctx, updateWidget, w.ID, w.HeaderTH, w.HeaderEN, w.SubHeaderTH, w.SubHeaderEN,
			w.StartDate, w.EndDate, w.ProductListRefID, w.UpdatedAt, w.UpdatedBy)
		if err != nil {
			return mapErr(err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("%w: widget %s", domain.ErrNotFound, w.ID)
		}
		if _, err := db.Exec(ctx, deleteComponents, w.ID); err != nil {
			return mapErr(err)
		}
		return r.insertComponents(ctx, db, w)
	})
}

// GetWidget implements port.WidgetRepository.
func (r *Repo) GetWidget(ctx context.Context, id string) (domain.Widget, error) {
	return r.get(ctx, selectWidget, id)
}

// GetWidgetForUpdate locks the row until the transaction in ctx ends.
func (r *Repo) GetWidgetForUpdate(ctx context.Context, id string) (domain.Widget, error) {
	if _, ok := txFrom(ctx); !ok {
		return domain.Widget{}, errors.New("postgres: GetWidgetForUpdate outside WithinTx")
	}
	return r.get(ctx, selectWidget+" FOR UPDATE", id)
}

func (r *Repo) get(ctx context.Context, q, id string) (domain.Widget, error) {
	db := conn(ctx, r.pool)
	w, err := scanWidget(db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Widget{}, fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.Widget{}, mapErr(err)
	}
	ws := []domain.Widget{w}
	if err := r.attachComponents(ctx, db, ws); err != nil {
		return domain.Widget{}, err
	}
	return ws[0], nil
}

// ListWidgets implements port.WidgetRepository.
func (r *Repo) ListWidgets(ctx context.Context, p domain.Page) ([]domain.Widget, int, error) {
	db := conn(ctx, r.pool)
	var total int
	if err := db.QueryRow(ctx, countWidgets).Scan(&total); err != nil {
		return nil, 0, mapErr(err)
	}
	rows, err := db.Query(ctx, selectWidgetPage, p.Size, p.Offset())
	if err != nil {
		return nil, 0, mapErr(err)
	}
	ws, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Widget, error) { return scanWidget(row) })
	if err != nil {
		return nil, 0, mapErr(err)
	}
	if err := r.attachComponents(ctx, db, ws); err != nil {
		return nil, 0, err
	}
	return ws, total, nil
}

// DeleteWidget implements port.WidgetRepository.
func (r *Repo) DeleteWidget(ctx context.Context, id string) error {
	tag, err := conn(ctx, r.pool).Exec(ctx, deleteWidget, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
	}
	return nil
}

func scanWidget(row pgx.Row) (domain.Widget, error) {
	var w domain.Widget
	err := row.Scan(&w.ID, &w.HeaderTH, &w.HeaderEN, &w.SubHeaderTH, &w.SubHeaderEN,
		&w.StartDate, &w.EndDate, &w.ProductListRefID, &w.CreatedAt, &w.CreatedBy, &w.UpdatedAt, &w.UpdatedBy)
	return w, err
}

func (r *Repo) insertComponents(ctx context.Context, db querier, w domain.Widget) error {
	for i, c := range w.Components {
		if _, err := db.Exec(ctx, insertComponent, w.ID, i, string(c.Type), c.SubType, c.ImageURL, c.IsActive, c.DataSrc); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func (r *Repo) attachComponents(ctx context.Context, db querier, ws []domain.Widget) error {
	if len(ws) == 0 {
		return nil
	}
	ids := make([]string, len(ws))
	byID := make(map[string]*domain.Widget, len(ws))
	for i := range ws {
		ids[i] = ws[i].ID
		ws[i].Components = []domain.Component{}
		byID[ws[i].ID] = &ws[i]
	}

	rows, err := db.Query(ctx, selectComponents, ids)
	if err != nil {
		return mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var widgetID, typ string
		var c domain.Component
		if err := rows.Scan(&widgetID, &typ, &c.SubType, &c.ImageURL, &c.IsActive, &c.DataSrc); err != nil {
			return mapErr(err)
		}
		c.Type = domain.ComponentType(typ)
		byID[widgetID].Components = append(byID[widgetID].Components, c)
	}
	return mapErr(rows.Err())
}
