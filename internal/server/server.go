// Package server is the composition root.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/config"
	"github.com/kritpi/scaffold-golang/internal/core/service"
	"github.com/kritpi/scaffold-golang/internal/handler"
	"github.com/kritpi/scaffold-golang/internal/httpx"
	"github.com/kritpi/scaffold-golang/internal/platform"
	"github.com/kritpi/scaffold-golang/internal/repository"
	"github.com/kritpi/scaffold-golang/internal/router"
)

// Run serves the API until ctx is cancelled, then drains in-flight requests for up to ShutdownTimeout.
func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := newLogger(cfg.Server)
	slog.SetDefault(log)

	verifier, err := auth.New(cfg.Auth)
	if err != nil {
		return err
	}

	repos, closeRepos, err := repository.New(ctx, cfg.Repository, log)
	if err != nil {
		return err
	}
	defer closeRepos()

	svcs := service.New(cfg.Service, service.Deps{Repositories: repos, Clock: platform.SystemClock{}})

	gin.SetMode(cfg.Server.GinMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), httpx.Deadline(cfg.Server.RequestTimeout))
	router.Setup(engine, handler.New(svcs, cfg.Handler), verifier)

	return serve(ctx, cfg.Server, engine, log)
}

func serve(ctx context.Context, cfg config.Server, h http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler: h,
	}

	errc := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	log.Info("shutting down", "timeout", cfg.ShutdownTimeout)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newLogger(cfg config.Server) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if cfg.LogJSON {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("service", cfg.ServiceName)
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts)).With("service", cfg.ServiceName)
}
