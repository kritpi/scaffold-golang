// Package catalog implements port.CatalogAdapter against the external catalog HTTP API.
package catalog

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// maxDrain caps how much of a response body is read. A larger body closes the connection instead.
const maxDrain = 64 << 10

// Config locates the catalog API. Timeout applies to each call.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client implements port.CatalogAdapter over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
}

// New builds a Client from cfg.
func New(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// ProductListExists reports whether the catalog has product list refID; ErrUnavailable on any answer but 200 or 404.
func (c *Client) ProductListExists(ctx context.Context, refID string) (bool, error) {
	u := c.cfg.BaseURL + "/v1/product-lists/" + url.PathEscape(refID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, fmt.Errorf("build catalog request: %w", err)
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("X-API-KEY", c.cfg.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: catalog: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrain)) // drain so the connection is reused

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("%w: catalog returned %d", domain.ErrUnavailable, resp.StatusCode)
	}
}

// AllowAll is a CatalogAdapter that reports every product list as existing.
type AllowAll struct{}

// ProductListExists implements port.CatalogAdapter.
func (AllowAll) ProductListExists(context.Context, string) (bool, error) { return true, nil }
