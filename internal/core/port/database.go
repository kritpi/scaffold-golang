package port

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Repo is the database: the state this service owns.
type Repo interface {
	Transactor
	WidgetRepository
}

// Transactor runs fn atomically. Repo calls inside fn must use the ctx fn receives;
// a WithinTx inside fn joins the outer transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// WidgetRepository stores widgets.
type WidgetRepository interface {
	// CreateWidget stores w. ErrConflict if w.ID exists.
	CreateWidget(ctx context.Context, w domain.Widget) error
	// UpdateWidget replaces w. ErrNotFound if missing.
	UpdateWidget(ctx context.Context, w domain.Widget) error
	// GetWidget returns the widget. ErrNotFound if missing.
	GetWidget(ctx context.Context, id string) (domain.Widget, error)
	// GetWidgetForUpdate is GetWidget, and it locks the widget until the
	// transaction in ctx ends. It fails outside WithinTx.
	GetWidgetForUpdate(ctx context.Context, id string) (domain.Widget, error)
	// ListWidgets returns one page ordered by UpdatedAt desc, then ID, and the total count.
	ListWidgets(ctx context.Context, page domain.Page) (items []domain.Widget, total int, err error)
	// DeleteWidget removes the widget. ErrNotFound if missing.
	DeleteWidget(ctx context.Context, id string) error
}
