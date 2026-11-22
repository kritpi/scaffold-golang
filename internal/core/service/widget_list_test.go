package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func TestWidgetList(t *testing.T) {
	ctx := context.Background()

	t.Run("pages newest first", func(t *testing.T) {
		f := newFixture(t)
		for range 5 {
			f.clock.now = f.clock.now.Add(time.Minute)
			_, err := f.newWidgetService().Create(ctx, validInput())
			require.NoError(t, err)
		}

		r, err := f.newWidgetService().List(ctx, domain.Page{Number: 2, Size: 2})
		require.NoError(t, err)
		assert.Equal(t, 5, r.Total)
		assert.Equal(t, 3, r.LastPage())
		require.Len(t, r.Items, 2)
		assert.Equal(t, "3", r.Items[0].ID, "newest first: 5,4 | 3,2 | 1")
		assert.Equal(t, domain.StatusActive, r.Items[0].Status)
	})

	t.Run("out-of-range paging is normalized", func(t *testing.T) {
		f := newFixture(t)

		r, err := f.newWidgetService().List(ctx, domain.Page{Number: 0, Size: 1000})
		require.NoError(t, err)
		assert.Equal(t, domain.Page{Number: 1, Size: domain.MaxPageSize}, r.Page)
		assert.Empty(t, r.Items)
		assert.Equal(t, 1, r.LastPage())
	})

	t.Run("repository failure is wrapped", func(t *testing.T) {
		f := newFixture(t)
		f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrUnavailable}

		_, err := f.newWidgetService().List(ctx, domain.Page{})
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		assert.ErrorContains(t, err, "list widgets")
	})
}
