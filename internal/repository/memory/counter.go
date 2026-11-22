package memory

import (
	"context"
	"strconv"
)

// NextWidgetID implements port.IDGenerator.
func (r *RedisRepo) NextWidgetID(context.Context) (string, error) {
	return strconv.FormatInt(r.n.Add(1), 10), nil
}
