// Package config loads the process configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/core/service"
	"github.com/kritpi/scaffold-golang/internal/core/service/helper"
	"github.com/kritpi/scaffold-golang/internal/handler"
	"github.com/kritpi/scaffold-golang/internal/repository"
	"github.com/kritpi/scaffold-golang/internal/repository/catalog"
	"github.com/kritpi/scaffold-golang/internal/repository/postgres"
	"github.com/kritpi/scaffold-golang/internal/repository/redis"
)

// Config is the whole process configuration, one section per consumer.
type Config struct {
	Server     Server
	Auth       auth.Config
	Handler    handler.Config
	Service    service.Config
	Repository repository.Config
}

// Server holds the HTTP listener and logging settings.
type Server struct {
	ServiceName     string
	Host            string
	Port            int
	ShutdownTimeout time.Duration
	LogJSON         bool
	GinMode         string
	RequestTimeout  time.Duration
}

// Load reads every key from the environment and reports all parse errors together.
func Load() (Config, error) {
	var errs []error
	e := env{errs: &errs}

	// Each request's ctx ends after requestTimeout. A cache eviction blocks fills
	// for twice that, so a read that raced the write cannot store the old widget.
	requestTimeout := e.positiveDur("REQUEST_TIMEOUT", 10*time.Second)

	cfg := Config{
		Server: Server{
			ServiceName:     e.str("SERVICE_NAME", "scaffold-api"),
			Host:            e.str("HOST", "0.0.0.0"),
			Port:            e.int("PORT", 8080),
			ShutdownTimeout: e.durOrSeconds("SHUTDOWN_TIMEOUT", 30*time.Second),
			LogJSON:         e.bool("LOG_JSON", true),
			GinMode:         e.str("GIN_MODE", "release"),
			RequestTimeout:  requestTimeout,
		},
		Auth: auth.Config{
			PublicKeyFile: e.required("AUTH_JWT_PUBLIC_KEY_FILE"),
			Issuer:        e.required("AUTH_JWT_ISSUER"),
			Audience:      e.required("AUTH_JWT_AUDIENCE"),
		},
		Handler: handler.Config{WireZone: e.location("API_TIME_ZONE", "Asia/Bangkok")},
		Service: service.Config{
			Widget: service.WidgetConfig{DataSrc: helper.WidgetDataSrc{
				ProductDataSrc:     e.template("WIDGET_PRODUCT_DATA_SRC", "/common-api/v2/widgets/%s/products?type=%s", 2),
				InformationSubType: e.str("WIDGET_PROMOTION_TYPE", "PROMOTION"),
				CategoryDataSrc:    e.template("WIDGET_CATEGORY_DATA_SRC", "/common-api/v1/widgets/%s/categories", 1),
			}},
		},
		Repository: repository.Config{
			EvictHold: 2 * requestTimeout,
			Postgres: postgres.Config{
				DSN:      e.str("POSTGRES_DSN", ""),
				MaxConns: int32(e.int("POSTGRES_MAX_CONNS", 10)),
			},
			Redis: redis.Config{
				Addr:     e.str("REDIS_ADDR", ""),
				Password: e.str("REDIS_PASSWORD", ""),
				DB:       e.int("REDIS_DB", 0),
				CacheTTL: e.positiveDur("WIDGET_CACHE_TTL", 5*time.Minute),
			},
			Catalog: catalog.Config{
				BaseURL: e.str("CATALOG_BASE_URL", ""),
				APIKey:  e.str("CATALOG_API_KEY", ""),
				Timeout: e.durOrSeconds("CATALOG_TIMEOUT", 3*time.Second),
			},
		},
	}
	if len(errs) > 0 {
		return Config{}, fmt.Errorf("config: %v", errs)
	}
	return cfg, nil
}

type env struct{ errs *[]error }

func (e env) str(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}

func (e env) template(k, def string, n int) string {
	v := e.str(k, def)
	if strings.Count(v, "%") != n || strings.Count(v, "%s") != n {
		*e.errs = append(*e.errs, fmt.Errorf("%s: want exactly %d %%s and no other verb, got %q", k, n, v))
	}
	return v
}

func (e env) location(k, def string) *time.Location {
	loc, err := time.LoadLocation(e.str(k, def))
	if err != nil {
		*e.errs = append(*e.errs, fmt.Errorf("%s: %w", k, err))
	}
	return loc
}

func (e env) required(k string) string {
	v := e.str(k, "")
	if v == "" {
		*e.errs = append(*e.errs, fmt.Errorf("%s: required", k))
	}
	return v
}

func (e env) positiveDur(k string, def time.Duration) time.Duration {
	d := e.durOrSeconds(k, def)
	if d <= 0 {
		*e.errs = append(*e.errs, fmt.Errorf("%s: must be positive, got %s", k, d))
	}
	return d
}

func (e env) int(k string, def int) int {
	v, ok := os.LookupEnv(k)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		*e.errs = append(*e.errs, fmt.Errorf("%s: %w", k, err))
	}
	return n
}

func (e env) bool(k string, def bool) bool {
	v, ok := os.LookupEnv(k)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		*e.errs = append(*e.errs, fmt.Errorf("%s: %w", k, err))
	}
	return b
}

func (e env) durOrSeconds(k string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(k)
	if !ok {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		*e.errs = append(*e.errs, fmt.Errorf("%s: %w", k, err))
	}
	return d
}
