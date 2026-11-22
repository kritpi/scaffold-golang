// Package memory implements the repository ports in process, for local runs and tests.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

type widgetRecord struct {
	ID               string
	HeaderTH         string
	HeaderEN         string
	SubHeaderTH      string
	SubHeaderEN      string
	StartDate        time.Time
	EndDate          time.Time
	ProductListRefID string
	Components       []componentRecord
	CreatedAt        time.Time
	CreatedBy        string
	UpdatedAt        time.Time
	UpdatedBy        string
}

type componentRecord struct {
	Type, SubType, ImageURL, DataSrc string
	IsActive                         bool
}

func toRecord(w domain.Widget) widgetRecord {
	cs := make([]componentRecord, len(w.Components))
	for i, c := range w.Components {
		cs[i] = componentRecord{Type: string(c.Type), SubType: c.SubType, ImageURL: c.ImageURL, DataSrc: c.DataSrc, IsActive: c.IsActive}
	}
	return widgetRecord{
		ID: w.ID, HeaderTH: w.HeaderTH, HeaderEN: w.HeaderEN, SubHeaderTH: w.SubHeaderTH, SubHeaderEN: w.SubHeaderEN,
		StartDate: w.StartDate, EndDate: w.EndDate, ProductListRefID: w.ProductListRefID, Components: cs,
		CreatedAt: w.CreatedAt, CreatedBy: w.CreatedBy, UpdatedAt: w.UpdatedAt, UpdatedBy: w.UpdatedBy,
	}
}

func (r widgetRecord) toDomain() domain.Widget {
	cs := make([]domain.Component, len(r.Components))
	for i, c := range r.Components {
		cs[i] = domain.Component{Type: domain.ComponentType(c.Type), SubType: c.SubType, ImageURL: c.ImageURL, DataSrc: c.DataSrc, IsActive: c.IsActive}
	}
	return domain.Widget{
		ID: r.ID, HeaderTH: r.HeaderTH, HeaderEN: r.HeaderEN, SubHeaderTH: r.SubHeaderTH, SubHeaderEN: r.SubHeaderEN,
		StartDate: r.StartDate, EndDate: r.EndDate, ProductListRefID: r.ProductListRefID, Components: cs,
		CreatedAt: r.CreatedAt, CreatedBy: r.CreatedBy, UpdatedAt: r.UpdatedAt, UpdatedBy: r.UpdatedBy,
	}
}

// Repo implements port.Repo with a map.
type Repo struct {
	mu   sync.RWMutex
	rows map[string]widgetRecord
	txMu sync.Mutex
}

// NewRepo returns an empty Repo.
func NewRepo() *Repo {
	return &Repo{rows: map[string]widgetRecord{}}
}

// CreateWidget implements port.WidgetRepository.
func (r *Repo) CreateWidget(ctx context.Context, w domain.Widget) error {
	return r.write(ctx, func() error {
		if _, ok := r.rows[w.ID]; ok {
			return fmt.Errorf("%w: widget %s already exists", domain.ErrConflict, w.ID)
		}
		r.rows[w.ID] = toRecord(w)
		return nil
	})
}

// UpdateWidget implements port.WidgetRepository.
func (r *Repo) UpdateWidget(ctx context.Context, w domain.Widget) error {
	return r.write(ctx, func() error {
		if _, ok := r.rows[w.ID]; !ok {
			return fmt.Errorf("%w: widget %s", domain.ErrNotFound, w.ID)
		}
		r.rows[w.ID] = toRecord(w)
		return nil
	})
}

// GetWidgetForUpdate implements port.WidgetRepository.
func (r *Repo) GetWidgetForUpdate(ctx context.Context, id string) (domain.Widget, error) {
	if !inTx(ctx) {
		return domain.Widget{}, errNoTx
	}
	return r.GetWidget(ctx, id)
}

// GetWidget implements port.WidgetRepository.
func (r *Repo) GetWidget(_ context.Context, id string) (domain.Widget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.rows[id]
	if !ok {
		return domain.Widget{}, fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
	}
	return rec.toDomain(), nil
}

// ListWidgets implements port.WidgetRepository.
func (r *Repo) ListWidgets(_ context.Context, p domain.Page) ([]domain.Widget, int, error) {
	r.mu.RLock()
	all := make([]widgetRecord, 0, len(r.rows))
	for _, rec := range r.rows {
		all = append(all, rec)
	}
	r.mu.RUnlock()

	sort.Slice(all, func(i, j int) bool {
		if !all[i].UpdatedAt.Equal(all[j].UpdatedAt) {
			return all[i].UpdatedAt.After(all[j].UpdatedAt)
		}
		return all[i].ID < all[j].ID
	})

	total := len(all)
	start := min(p.Offset(), total)
	end := min(start+p.Size, total)
	out := make([]domain.Widget, 0, end-start)
	for _, rec := range all[start:end] {
		out = append(out, rec.toDomain())
	}
	return out, total, nil
}

// DeleteWidget implements port.WidgetRepository.
func (r *Repo) DeleteWidget(ctx context.Context, id string) error {
	return r.write(ctx, func() error {
		if _, ok := r.rows[id]; !ok {
			return fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
		}
		delete(r.rows, id)
		return nil
	})
}
