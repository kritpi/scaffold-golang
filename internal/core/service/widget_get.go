package service

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Get reads the widget through the cache and derives its status now.
func (s *widgetService) Get(ctx context.Context, id string) (domain.WidgetView, error) {
	w, ok := s.redisRepo.GetCachedWidget(ctx, id)
	if !ok {
		var err error
		if w, err = s.repo.GetWidget(ctx, id); err != nil {
			return domain.WidgetView{}, fmt.Errorf("get widget %s: %w", id, err)
		}
		s.redisRepo.CacheWidget(ctx, w)
	}
	return domain.WidgetView{Widget: w, Status: w.StatusAt(s.clock.Now())}, nil
}
