package service

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// List returns one normalized page with each status derived now.
func (s *widgetService) List(ctx context.Context, page domain.Page) (domain.PageResult[domain.WidgetView], error) {
	page = page.Normalize()
	ws, total, err := s.repo.ListWidgets(ctx, page)
	if err != nil {
		return domain.PageResult[domain.WidgetView]{}, fmt.Errorf("list widgets: %w", err)
	}

	now := s.clock.Now()
	views := make([]domain.WidgetView, len(ws))
	for i, w := range ws {
		views[i] = domain.WidgetView{Widget: w, Status: w.StatusAt(now)}
	}
	return domain.PageResult[domain.WidgetView]{Items: views, Page: page, Total: total}, nil
}
