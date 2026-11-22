package service

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
	"github.com/kritpi/scaffold-golang/internal/core/service/helper"
)

// Create validates in, allocates an ID, stores the widget and publishes widget.created.
func (s *widgetService) Create(ctx context.Context, in domain.WidgetInput) (string, error) {
	if err := helper.CheckWidgetInput(ctx, s.adpt, in); err != nil {
		return "", err
	}

	id, err := s.redisRepo.NextWidgetID(ctx)
	if err != nil {
		return "", fmt.Errorf("allocate widget id: %w", err)
	}

	now := s.clock.Now()
	in.Components = helper.ResolveWidgetDataSrc(s.cfg.DataSrc, id, in.Components)
	w, err := domain.NewWidget(id, in, now)
	if err != nil {
		return "", err
	}

	if err := s.repo.CreateWidget(ctx, w); err != nil {
		return "", fmt.Errorf("create widget %s: %w", id, err)
	}

	s.redisRepo.PublishWidgetEvent(context.WithoutCancel(ctx), domain.WidgetEvent{Type: domain.EventWidgetCreated, WidgetID: id, Actor: in.Actor, At: now})
	return id, nil
}
