// Package httpx adapts typed handler funcs to Gin.
package httpx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Handle turns a typed func(ctx, Req) (Resp, error) into a gin.HandlerFunc:
// bind URI + query + JSON body into Req, call fn, write Resp with status or
// map the error.
func Handle[Req, Resp any](status int, fn func(context.Context, Req) (Resp, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := bind(c, &req); err != nil {
			WriteError(c, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
			return
		}

		resp, err := fn(c.Request.Context(), req)
		if err != nil {
			WriteError(c, err)
			return
		}
		c.JSON(status, resp)
	}
}

func bind(c *gin.Context, req any) error {
	if len(c.Params) > 0 {
		if err := c.ShouldBindUri(req); err != nil {
			return err
		}
	}
	if c.Request.URL.RawQuery != "" {
		if err := c.ShouldBindQuery(req); err != nil {
			return err
		}
	}
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if c.Request.ContentLength != 0 {
			if err := c.ShouldBindJSON(req); err != nil {
				return err
			}
		}
	}
	return nil
}
