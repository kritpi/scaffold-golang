package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// WriteError logs err and writes the error envelope with the status of its domain kind.
func WriteError(c *gin.Context, err error) {
	status, code := classify(err)
	ctx := c.Request.Context()
	attrs := []any{"method", c.Request.Method, "path", c.FullPath(), "status", status, "err", err}
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "request failed", attrs...)
	} else {
		slog.InfoContext(ctx, "request rejected", attrs...)
	}

	msg := err.Error()
	if status >= http.StatusInternalServerError {
		msg = strings.ToLower(http.StatusText(status))
	}
	c.AbortWithStatusJSON(status, dto.ErrorResponse{Code: code, Message: msg})
}

func classify(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, "INVALID_REQUEST"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable, "UNAVAILABLE"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}
