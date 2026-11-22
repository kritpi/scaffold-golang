package service

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
	"github.com/kritpi/scaffold-golang/internal/core/service/helper"
)

// Update revises the widget in a transaction, then evicts its cache entry and publishes widget.updated.
func (s *widgetService) Update(ctx context.Context, id string, in domain.WidgetInput) error {
	if err := helper.CheckWidgetInput(ctx, s.adpt, in); err != nil {
		return err
	}

	now := s.clock.Now()
	err := s.repo.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.repo.GetWidgetForUpdate(ctx, id)
		if err != nil {
			return err
		}
		in.Components = helper.ResolveWidgetDataSrc(s.cfg.DataSrc, id, in.Components)
		next, err := current.Revise(in, now)
		if err != nil {
			return err
		}
		return s.repo.UpdateWidget(ctx, next)
	})
	if err != nil {
		return fmt.Errorf("update widget %s: %w", id, err)
	}

	committed := context.WithoutCancel(ctx)
	s.redisRepo.EvictWidget(committed, id)
	s.redisRepo.PublishWidgetEvent(committed, domain.WidgetEvent{Type: domain.EventWidgetUpdated, WidgetID: id, Actor: in.Actor, At: now})
	return nil
}
