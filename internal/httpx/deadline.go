package httpx

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Deadline ends every request's ctx after d.
func Deadline(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
