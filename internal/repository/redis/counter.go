package redis

import (
	"context"
	"strconv"
)

// NextWidgetID implements port.IDGenerator.
func (c *Repo) NextWidgetID(ctx context.Context) (string, error) {
	n, err := c.rdb.Incr(ctx, widgetIDKey).Result()
	if err != nil {
		return "", unavailable(err)
	}
	return strconv.FormatInt(n, 10), nil
}
