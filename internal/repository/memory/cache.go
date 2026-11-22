package memory

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// GetCachedWidget implements port.WidgetCache.
func (c *RedisRepo) GetCachedWidget(_ context.Context, id string) (domain.Widget, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rec, ok := c.cache[id]
	return rec.toDomain(), ok
}

// CacheWidget implements port.WidgetCache.
func (c *RedisRepo) CacheWidget(ctx context.Context, w domain.Widget) {
	if ctx.Err() != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, filled := c.cache[w.ID]
	if filled || c.clock.Now().Before(c.evicted[w.ID]) {
		return
	}
	c.cache[w.ID] = toRecord(w)
}

// EvictWidget implements port.WidgetCache.
func (c *RedisRepo) EvictWidget(_ context.Context, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, id)
	c.evicted[id] = c.clock.Now().Add(c.hold)
}
