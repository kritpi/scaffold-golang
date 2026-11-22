package port

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Services bundles the driving ports that service.New builds.
type Services struct {
	Widget WidgetService
}

// WidgetService is the widget use-case port.
type WidgetService interface {
	Create(ctx context.Context, in domain.WidgetInput) (string, error)
	Update(ctx context.Context, id string, in domain.WidgetInput) error
	Get(ctx context.Context, id string) (domain.WidgetView, error)
	List(ctx context.Context, page domain.Page) (domain.PageResult[domain.WidgetView], error)
	Delete(ctx context.Context, id string) error
}
