package port

import (
	"context"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// RedisRepo is the key-value store: IDs, the read cache and the event stream.
type RedisRepo interface {
	IDGenerator
	WidgetCache
	EventPublisher
}

// Adapter is every external API this service calls.
type Adapter interface {
	CatalogAdapter
}

// IDGenerator hands out widget IDs that no stored widget already has.
type IDGenerator interface {
	NextWidgetID(ctx context.Context) (string, error)
}

// WidgetCache is best effort. Implementations log failures and report them as misses.
// CacheWidget fills only an absent entry. EvictWidget blocks fills for a hold
// window, so a read that raced a write cannot store the old widget.
type WidgetCache interface {
	GetCachedWidget(ctx context.Context, id string) (domain.Widget, bool)
	CacheWidget(ctx context.Context, w domain.Widget)
	EvictWidget(ctx context.Context, id string)
}

// EventPublisher is fire-and-forget. Implementations log their own failures.
type EventPublisher interface {
	PublishWidgetEvent(ctx context.Context, e domain.WidgetEvent)
}

// CatalogAdapter asks the external catalog service about product lists.
// ErrUnavailable when the catalog cannot answer.
type CatalogAdapter interface {
	ProductListExists(ctx context.Context, refID string) (bool, error)
}

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}
