package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func TestWidgetCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success persists, resolves data src, emits event", func(t *testing.T) {
		f := newFixture(t)

		id, err := f.newWidgetService().Create(ctx, validInput())
		require.NoError(t, err)
		assert.Equal(t, "1", id)

		got, err := f.repo.GetWidget(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "/p/1?type=IA", got.Components[0].DataSrc)
		assert.Equal(t, "/p/1?type=PROMOTION", got.Components[1].DataSrc)
		assert.Equal(t, "/c/1", got.Components[2].DataSrc)
		assert.Equal(t, f.clock.now, got.CreatedAt)
		assert.Equal(t, "ops@example.com", got.CreatedBy)

		require.Len(t, f.events.got, 1)
		assert.Equal(t, domain.WidgetEvent{Type: domain.EventWidgetCreated, WidgetID: "1", Actor: "ops@example.com", At: f.clock.now}, f.events.got[0])
	})

	t.Run("invalid input is rejected before any side effect", func(t *testing.T) {
		cases := map[string]func(*domain.WidgetInput){
			"missing actor":          func(in *domain.WidgetInput) { in.Actor = "" },
			"missing header":         func(in *domain.WidgetInput) { in.HeaderTH = "" },
			"missing start date":     func(in *domain.WidgetInput) { in.StartDate = time.Time{} },
			"end before start":       func(in *domain.WidgetInput) { in.StartDate, in.EndDate = in.EndDate, in.StartDate },
			"unknown component type": func(in *domain.WidgetInput) { in.Components[0].Type = "BANNER" },
		}
		for name, mutate := range cases {
			t.Run(name, func(t *testing.T) {
				f := newFixture(t)
				in := validInput()
				mutate(&in)

				_, err := f.newWidgetService().Create(ctx, in)
				assert.ErrorIs(t, err, domain.ErrInvalidInput)

				assert.Zero(t, f.catalog.calls, "catalog is not called")
				next, _ := f.deps.RedisRepo.NextWidgetID(ctx)
				assert.Equal(t, "1", next, "no ID is consumed")
				assert.Empty(t, f.events.got)
			})
		}
	})

	t.Run("sub type is query-escaped into data src", func(t *testing.T) {
		f := newFixture(t)
		in := validInput()
		in.Components[0].SubType = "IA&admin=1"

		id, err := f.newWidgetService().Create(ctx, in)
		require.NoError(t, err)
		got, _ := f.repo.GetWidget(ctx, id)
		assert.Equal(t, "/p/1?type=IA%26admin%3D1", got.Components[0].DataSrc)
	})

	t.Run("unknown product list is invalid input", func(t *testing.T) {
		f := newFixture(t)
		in := validInput()
		in.ProductListRefID = "nope"

		_, err := f.newWidgetService().Create(ctx, in)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("empty product list skips the catalog", func(t *testing.T) {
		f := newFixture(t)
		f.catalog.err = domain.ErrUnavailable
		in := validInput()
		in.ProductListRefID = ""

		_, err := f.newWidgetService().Create(ctx, in)
		require.NoError(t, err)
		assert.Zero(t, f.catalog.calls)
	})

	t.Run("catalog outage is unavailable and consumes no ID", func(t *testing.T) {
		f := newFixture(t)
		f.catalog.err = domain.ErrUnavailable

		_, err := f.newWidgetService().Create(ctx, validInput())
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		next, _ := f.deps.RedisRepo.NextWidgetID(ctx)
		assert.Equal(t, "1", next)
	})

	t.Run("id generator failure is wrapped", func(t *testing.T) {
		f := newFixture(t)
		f.deps.RedisRepo = failingIDs{RedisRepo: f.events, err: domain.ErrUnavailable}

		_, err := f.newWidgetService().Create(ctx, validInput())
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		assert.ErrorContains(t, err, "allocate widget id")
		assert.Empty(t, f.events.got)
	})

	t.Run("repository failure is wrapped and emits nothing", func(t *testing.T) {
		f := newFixture(t)
		f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrConflict}

		_, err := f.newWidgetService().Create(ctx, validInput())
		assert.ErrorIs(t, err, domain.ErrConflict)
		assert.ErrorContains(t, err, "create widget 1")
		assert.Empty(t, f.events.got)
	})
}
