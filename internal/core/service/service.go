// Package service implements the driving ports declared in package port.
package service

import (
	"github.com/kritpi/scaffold-golang/internal/core/port"
	"github.com/kritpi/scaffold-golang/internal/core/service/helper"
)

// Config holds the business settings, one section per resource.
type Config struct {
	Widget WidgetConfig
}

// WidgetConfig holds the widget's business settings.
type WidgetConfig struct {
	DataSrc helper.WidgetDataSrc
}

// Deps is shared by every resource's service: one port per datasource plus the clock.
type Deps struct {
	port.Repositories
	Clock port.Clock
}

type deps struct {
	repo      port.Repo
	redisRepo port.RedisRepo
	adpt      port.Adapter
	clock     port.Clock
}

// New builds every service from one Config and one Deps.
func New(cfg Config, d Deps) port.Services {
	shared := deps{repo: d.Repo, redisRepo: d.RedisRepo, adpt: d.Adpt, clock: d.Clock}
	return port.Services{
		Widget: &widgetService{cfg: cfg.Widget, deps: shared},
	}
}

type widgetService struct {
	cfg WidgetConfig
	deps
}
