package service

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Delete removes the widget, evicts its cache entry and publishes widget.deleted.
func (s *widgetService) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteWidget(ctx, id); err != nil {
		return fmt.Errorf("delete widget %s: %w", id, err)
	}
	committed := context.WithoutCancel(ctx)
	s.redisRepo.EvictWidget(committed, id)
	s.redisRepo.PublishWidgetEvent(committed, domain.WidgetEvent{Type: domain.EventWidgetDeleted, WidgetID: id, At: s.clock.Now()})
	return nil
}
