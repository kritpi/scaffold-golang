// Package repository builds port.Repositories from Config. An empty section
// selects the in-process stand-in for its ports.
package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/port"
	"github.com/kritpi/scaffold-golang/internal/platform"
	"github.com/kritpi/scaffold-golang/internal/repository/catalog"
	"github.com/kritpi/scaffold-golang/internal/repository/memory"
	"github.com/kritpi/scaffold-golang/internal/repository/postgres"
	"github.com/kritpi/scaffold-golang/internal/repository/redis"
)

// Config holds one section per datasource. EvictHold is how long a cache
// eviction blocks fills.
type Config struct {
	EvictHold time.Duration
	Postgres  postgres.Config
	Redis     redis.Config
	Catalog   catalog.Config
}

// New opens every client it needs; close releases them.
func New(ctx context.Context, cfg Config, log *slog.Logger) (r port.Repositories, close func(), err error) {
	if cfg.Postgres.DSN != "" && cfg.Redis.Addr == "" {
		return port.Repositories{}, nil, errors.New("repository: POSTGRES_DSN needs REDIS_ADDR: the memory ID counter restarts at 1 and collides with stored widgets")
	}

	var closers []func()
	close = func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}

	r = port.Repositories{
		Repo:      memory.NewRepo(),
		RedisRepo: memory.NewRedisRepo(log, cfg.EvictHold, platform.SystemClock{}),
		Adpt:      catalog.AllowAll{},
	}

	if cfg.Postgres.DSN != "" {
		pool, err := postgres.Open(ctx, cfg.Postgres)
		if err != nil {
			close()
			return port.Repositories{}, nil, err
		}
		closers = append(closers, pool.Close)
		r.Repo = postgres.New(pool)
	}

	if cfg.Redis.Addr != "" {
		rdb, err := redis.Open(ctx, cfg.Redis)
		if err != nil {
			close()
			return port.Repositories{}, nil, err
		}
		closers = append(closers, func() { _ = rdb.Close() })
		r.RedisRepo = redis.New(rdb, cfg.Redis.CacheTTL, cfg.EvictHold, log)
	}

	if cfg.Catalog.BaseURL != "" {
		r.Adpt = catalog.New(cfg.Catalog)
	}
	return r, close, nil
}
