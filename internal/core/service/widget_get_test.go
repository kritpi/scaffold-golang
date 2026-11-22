package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func TestWidgetGet(t *testing.T) {
	ctx := context.Background()

	t.Run("status is derived from the schedule at read time", func(t *testing.T) {
		in := validInput()
		cases := []struct {
			name string
			now  time.Time
			want domain.Status
		}{
			{"before start", in.StartDate.Add(-time.Second), domain.StatusNew},
			{"exactly start", in.StartDate, domain.StatusActive},
			{"exactly end", in.EndDate, domain.StatusActive},
			{"after end", in.EndDate.Add(time.Second), domain.StatusInactive},
		}
		f := newFixture(t)
		id, err := f.newWidgetService().Create(ctx, in)
		require.NoError(t, err)
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f.clock.now = tc.now
				v, err := f.newWidgetService().Get(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, tc.want, v.Status)
			})
		}
	})

	t.Run("repeat reads are served from the cache", func(t *testing.T) {
		f := newFixture(t)
		svc := f.newWidgetService()
		id, _ := svc.Create(ctx, validInput())
		first, err := svc.Get(ctx, id)
		require.NoError(t, err)

		changed := first.Widget
		changed.HeaderTH = "written behind the service"
		require.NoError(t, f.repo.UpdateWidget(ctx, changed))

		v, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "หัวข้อ", v.HeaderTH)
	})

	t.Run("a read that races an update cannot cache the old widget", func(t *testing.T) {
		f := newFixture(t)
		id, err := f.newWidgetService().Create(ctx, validInput())
		require.NoError(t, err)
		race := &racingRepo{Repo: f.repo}
		f.deps.Repo = race
		svc := f.newWidgetService()
		race.beforeReturn = func() {
			in := validInput()
			in.HeaderTH = "ใหม่"
			require.NoError(t, svc.Update(ctx, id, in))
		}

		_, err = svc.Get(ctx, id)
		require.NoError(t, err)

		v, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "ใหม่", v.HeaderTH)
	})

	t.Run("after the eviction hold, reads are cached again", func(t *testing.T) {
		f := newFixture(t)
		svc := f.newWidgetService()
		id, _ := svc.Create(ctx, validInput())
		require.NoError(t, svc.Update(ctx, id, validInput()))

		f.clock.now = f.clock.now.Add(21 * time.Second)
		first, err := svc.Get(ctx, id)
		require.NoError(t, err)
		changed := first.Widget
		changed.HeaderTH = "written behind the service"
		require.NoError(t, f.repo.UpdateWidget(ctx, changed))

		v, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "หัวข้อ", v.HeaderTH)
	})

	t.Run("a read whose ctx ended does not fill the cache", func(t *testing.T) {
		f := newFixture(t)
		svc := f.newWidgetService()
		id, _ := svc.Create(ctx, validInput())
		ended, cancel := context.WithCancel(ctx)
		cancel()
		_, err := svc.Get(ended, id)
		require.NoError(t, err)

		w, _ := f.repo.GetWidget(ctx, id)
		w.HeaderTH = "written behind the service"
		require.NoError(t, f.repo.UpdateWidget(ctx, w))

		v, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "written behind the service", v.HeaderTH)
	})

	t.Run("missing widget is not found", func(t *testing.T) {
		f := newFixture(t)

		_, err := f.newWidgetService().Get(ctx, "404")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("repository failure is wrapped", func(t *testing.T) {
		f := newFixture(t)
		f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrUnavailable}

		_, err := f.newWidgetService().Get(ctx, "1")
		assert.ErrorIs(t, err, domain.ErrUnavailable)
		assert.ErrorContains(t, err, "get widget 1")
	})
}
