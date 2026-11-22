package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func TestWidgetDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("removes the widget, its cache entry, and emits event", func(t *testing.T) {
		f := newFixture(t)
		svc := f.newWidgetService()
		id, _ := svc.Create(ctx, validInput())
		_, err := svc.Get(ctx, id)
		require.NoError(t, err)

		require.NoError(t, svc.Delete(ctx, id))

		_, err = svc.Get(ctx, id)
		assert.ErrorIs(t, err, domain.ErrNotFound, "the cached copy is gone too")
		require.Len(t, f.events.got, 2)
		assert.Equal(t, domain.WidgetEvent{Type: domain.EventWidgetDeleted, WidgetID: id, At: f.clock.now}, f.events.got[1])
	})

	t.Run("missing widget is not found and emits nothing", func(t *testing.T) {
		f := newFixture(t)

		assert.ErrorIs(t, f.newWidgetService().Delete(ctx, "404"), domain.ErrNotFound)
		assert.Empty(t, f.events.got)
	})

	t.Run("repository failure is wrapped", func(t *testing.T) {
		f := newFixture(t)
		f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrUnavailable}

		err := f.newWidgetService().Delete(ctx, "1")
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		assert.ErrorContains(t, err, "delete widget 1")
	})
}
