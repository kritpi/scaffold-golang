package httpx

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Authenticate requires a valid bearer token and puts its email in the request ctx.
func Authenticate(v *auth.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		email, err := "", fmt.Errorf("%w: missing bearer token", domain.ErrUnauthenticated)
		if ok {
			email, err = v.Actor(token)
		}
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			WriteError(c, err)
			return
		}
		c.Request = c.Request.WithContext(auth.WithActor(c.Request.Context(), email))
		c.Next()
	}
}
