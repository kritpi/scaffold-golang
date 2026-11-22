// Package router maps URLs to handler methods.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/handler"
	"github.com/kritpi/scaffold-golang/internal/httpx"
)

// BasePath prefixes every API route.
const BasePath = "/scaffold-api"

// Setup registers every route.
func Setup(r *gin.Engine, h *handler.Handler, v *auth.Verifier) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group(BasePath+"/v1", httpx.Authenticate(v))

	widgets := v1.Group("/widgets")
	{
		widgets.GET("", httpx.Handle(http.StatusOK, h.ListWidgets))
		widgets.POST("", httpx.Handle(http.StatusCreated, h.CreateWidget))
		widgets.GET("/:widget-id", httpx.Handle(http.StatusOK, h.GetWidget))
		widgets.PUT("/:widget-id", httpx.Handle(http.StatusOK, h.UpdateWidget))
		widgets.DELETE("/:widget-id", httpx.Handle(http.StatusOK, h.DeleteWidget))
	}
}
