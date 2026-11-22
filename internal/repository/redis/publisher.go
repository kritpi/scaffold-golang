package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// PublishWidgetEvent implements port.EventPublisher.
func (p *Repo) PublishWidgetEvent(ctx context.Context, e domain.WidgetEvent) {
	err := p.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: widgetStream,
		MaxLen: streamMaxLen,
		Approx: true,
		Values: map[string]any{
			"type":      string(e.Type),
			"widget_id": e.WidgetID,
			"actor":     e.Actor,
			"at":        e.At.UTC().Format(time.RFC3339Nano),
		},
	}).Err()
	if err != nil {
		p.log.WarnContext(ctx, "publish widget event failed", "type", e.Type, "widget_id", e.WidgetID, "err", err)
	}
}
