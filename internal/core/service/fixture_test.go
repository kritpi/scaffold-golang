package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
	"github.com/kritpi/scaffold-golang/internal/core/port"
	"github.com/kritpi/scaffold-golang/internal/core/service"
	"github.com/kritpi/scaffold-golang/internal/core/service/helper"
	"github.com/kritpi/scaffold-golang/internal/repository/memory"
)

type fixture struct {
	cfg     service.Config
	deps    service.Deps
	repo    *memory.Repo
	events  *recordingRedis
	clock   *fakeClock
	catalog *fakeCatalog
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		cfg: service.Config{Widget: service.WidgetConfig{DataSrc: helper.WidgetDataSrc{
			ProductDataSrc: "/p/%s?type=%s", InformationSubType: "PROMOTION", CategoryDataSrc: "/c/%s",
		}}},
		repo:    memory.NewRepo(),
		clock:   &fakeClock{now: time.Date(2025, 12, 25, 12, 0, 0, 0, time.UTC)},
		catalog: &fakeCatalog{known: map[string]bool{"PL-1": true}},
	}
	f.events = &recordingRedis{RedisRepo: memory.NewRedisRepo(slog.New(slog.NewTextHandler(io.Discard, nil)), 20*time.Second, f.clock)}
	f.deps = service.Deps{
		Repositories: port.Repositories{Repo: f.repo, RedisRepo: f.events, Adpt: f.catalog},
		Clock:        f.clock,
	}
	return f
}

func (f *fixture) newWidgetService() port.WidgetService { return service.New(f.cfg, f.deps).Widget }

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type fakeCatalog struct {
	known map[string]bool
	err   error
	calls int
}

func (c *fakeCatalog) ProductListExists(_ context.Context, id string) (bool, error) {
	c.calls++
	return c.known[id], c.err
}

// recordingRedis is the memory RedisRepo, except that it records events.
type recordingRedis struct {
	port.RedisRepo
	got []domain.WidgetEvent
}

func (r *recordingRedis) PublishWidgetEvent(_ context.Context, e domain.WidgetEvent) {
	r.got = append(r.got, e)
}

// failingIDs is a RedisRepo whose NextWidgetID fails with err.
type failingIDs struct {
	port.RedisRepo
	err error
}

func (f failingIDs) NextWidgetID(context.Context) (string, error) { return "", f.err }

// failingRepo is a Repo whose widget methods fail with err. WithinTx still runs fn.
type failingRepo struct {
	port.Repo
	err error
}

func (f failingRepo) CreateWidget(context.Context, domain.Widget) error { return f.err }
func (f failingRepo) UpdateWidget(context.Context, domain.Widget) error { return f.err }
func (f failingRepo) DeleteWidget(context.Context, string) error        { return f.err }
func (f failingRepo) GetWidget(context.Context, string) (domain.Widget, error) {
	return domain.Widget{}, f.err
}
func (f failingRepo) GetWidgetForUpdate(context.Context, string) (domain.Widget, error) {
	return domain.Widget{}, f.err
}
func (f failingRepo) ListWidgets(context.Context, domain.Page) ([]domain.Widget, int, error) {
	return nil, 0, f.err
}

// racingRepo runs beforeReturn once, after GetWidget has read the row and
// before it returns, so a test can commit a write in that window.
type racingRepo struct {
	port.Repo
	beforeReturn func()
}

func (r *racingRepo) GetWidget(ctx context.Context, id string) (domain.Widget, error) {
	w, err := r.Repo.GetWidget(ctx, id)
	if h := r.beforeReturn; h != nil {
		r.beforeReturn = nil
		h()
	}
	return w, err
}
