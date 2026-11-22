// Package redis implements port.RedisRepo on Redis.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

const (
	widgetIDKey          = "scaffold:widget-id"
	widgetCacheKeyPrefix = "scaffold:widget-cache:"
	widgetStream         = "scaffold:widget-events"
	streamMaxLen         = 10000

	evictedMarker = "evicted"
)

// Config selects the Redis server and the cache TTL.
type Config struct {
	Addr     string
	Password string
	DB       int
	CacheTTL time.Duration
}

// Repo implements port.RedisRepo on Redis.
type Repo struct {
	rdb  *goredis.Client
	ttl  time.Duration
	hold time.Duration
	log  *slog.Logger
}

// New wraps an open client. ttl applies to cached widgets; EvictWidget blocks fills for hold.
func New(rdb *goredis.Client, ttl, hold time.Duration, log *slog.Logger) *Repo {
	return &Repo{rdb: rdb, ttl: ttl, hold: hold, log: log}
}

// Open pings the server.
func Open(ctx context.Context, cfg Config) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis %s: %w", cfg.Addr, err)
	}
	return rdb, nil
}

func unavailable(err error) error {
	return fmt.Errorf("%w: redis: %v", domain.ErrUnavailable, err)
}
