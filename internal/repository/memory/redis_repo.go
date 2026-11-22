package memory

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/port"
)

// RedisRepo implements port.RedisRepo in process.
type RedisRepo struct {
	n       atomic.Int64
	mu      sync.RWMutex
	cache   map[string]widgetRecord
	evicted map[string]time.Time // id -> end of its hold
	hold    time.Duration
	clock   port.Clock
	log     *slog.Logger
}

// NewRedisRepo returns an empty RedisRepo. EvictWidget blocks fills for hold.
func NewRedisRepo(log *slog.Logger, hold time.Duration, clock port.Clock) *RedisRepo {
	return &RedisRepo{cache: map[string]widgetRecord{}, evicted: map[string]time.Time{}, hold: hold, clock: clock, log: log}
}
