// Package handler maps HTTP DTOs to port.Services calls.
package handler

import (
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/port"
)

// Config is the wire settings every endpoint shares.
type Config struct {
	WireZone *time.Location
}

// Handler serves every resource from the one port.Services value.
type Handler struct {
	svc port.Services
	loc *time.Location
}

// New builds the Handler for every resource.
func New(svc port.Services, cfg Config) *Handler { return &Handler{svc: svc, loc: cfg.WireZone} }
