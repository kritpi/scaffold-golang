package memory

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// PublishWidgetEvent implements port.EventPublisher.
func (r *RedisRepo) PublishWidgetEvent(ctx context.Context, e domain.WidgetEvent) {
	r.log.InfoContext(ctx, "event", "type", e.Type, "widget_id", e.WidgetID, "actor", e.Actor, "at", e.At)
}
