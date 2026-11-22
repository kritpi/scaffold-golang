package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func TestWidgetUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("keeps created fields, bumps updated fields, emits event", func(t *testing.T) {
		f := newFixture(t)
		created := f.clock.now
		id, err := f.newWidgetService().Create(ctx, validInput())
		require.NoError(t, err)

		f.clock.now = f.clock.now.Add(time.Hour)
		in := validInput()
		in.Actor, in.HeaderTH = "editor@example.com", "ใหม่"
		require.NoError(t, f.newWidgetService().Update(ctx, id, in))

		got, err := f.repo.GetWidget(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "ops@example.com", got.CreatedBy)
		assert.Equal(t, created, got.CreatedAt)
		assert.Equal(t, "editor@example.com", got.UpdatedBy)
		assert.Equal(t, f.clock.now, got.UpdatedAt)
		assert.Equal(t, "ใหม่", got.HeaderTH)
		assert.Equal(t, "/p/1?type=IA", got.Components[0].DataSrc)

		require.Len(t, f.events.got, 2)
		assert.Equal(t, domain.WidgetEvent{Type: domain.EventWidgetUpdated, WidgetID: id, Actor: "editor@example.com", At: f.clock.now}, f.events.got[1])
	})

	t.Run("a cached read sees the update", func(t *testing.T) {
		f := newFixture(t)
		svc := f.newWidgetService()
		id, _ := svc.Create(ctx, validInput())
		_, err := svc.Get(ctx, id)
		require.NoError(t, err)

		in := validInput()
		in.HeaderTH = "ใหม่"
		require.NoError(t, svc.Update(ctx, id, in))

		v, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "ใหม่", v.HeaderTH)
	})

	t.Run("missing widget is not found and emits nothing", func(t *testing.T) {
		f := newFixture(t)

		err := f.newWidgetService().Update(ctx, "404", validInput())
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Empty(t, f.events.got)
	})

	t.Run("invalid input leaves the widget unchanged", func(t *testing.T) {
		f := newFixture(t)
		id, _ := f.newWidgetService().Create(ctx, validInput())
		in := validInput()
		in.HeaderTH = "ใหม่"
		in.StartDate, in.EndDate = in.EndDate, in.StartDate

		assert.ErrorIs(t, f.newWidgetService().Update(ctx, id, in), domain.ErrInvalidInput)
		got, _ := f.repo.GetWidget(ctx, id)
		assert.Equal(t, "หัวข้อ", got.HeaderTH)
	})

	t.Run("catalog outage is unavailable", func(t *testing.T) {
		f := newFixture(t)
		id, _ := f.newWidgetService().Create(ctx, validInput())
		f.catalog.err = domain.ErrUnavailable

		assert.ErrorIs(t, f.newWidgetService().Update(ctx, id, validInput()), domain.ErrUnavailable)
	})

	t.Run("repository failure is wrapped", func(t *testing.T) {
		f := newFixture(t)
		f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrUnavailable}

		err := f.newWidgetService().Update(ctx, "1", validInput())
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		assert.ErrorContains(t, err, "update widget 1")
	})
}
