package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// GetCachedWidget implements port.WidgetCache.
func (c *Repo) GetCachedWidget(ctx context.Context, id string) (domain.Widget, bool) {
	b, err := c.rdb.Get(ctx, widgetCacheKeyPrefix+id).Bytes()
	if errors.Is(err, goredis.Nil) || string(b) == evictedMarker {
		return domain.Widget{}, false
	}
	var rec widgetRecord
	if err == nil {
		err = json.Unmarshal(b, &rec)
	}
	if err != nil {
		c.log.WarnContext(ctx, "widget cache get failed", "widget_id", id, "err", err)
		return domain.Widget{}, false
	}
	return rec.toDomain(), true
}

// CacheWidget implements port.WidgetCache.
func (c *Repo) CacheWidget(ctx context.Context, w domain.Widget) {
	if ctx.Err() != nil {
		return
	}
	b, err := json.Marshal(toRecord(w))
	if err == nil {
		err = c.rdb.SetNX(ctx, widgetCacheKeyPrefix+w.ID, b, c.ttl).Err()
	}
	if err != nil {
		c.log.WarnContext(ctx, "widget cache set failed", "widget_id", w.ID, "err", err)
	}
}

// EvictWidget implements port.WidgetCache.
func (c *Repo) EvictWidget(ctx context.Context, id string) {
	if err := c.rdb.Set(ctx, widgetCacheKeyPrefix+id, evictedMarker, c.hold).Err(); err != nil {
		c.log.WarnContext(ctx, "widget cache evict failed", "widget_id", id, "err", err)
	}
}

type widgetRecord struct {
	ID               string            `json:"id"`
	HeaderTH         string            `json:"header_th"`
	HeaderEN         string            `json:"header_en"`
	SubHeaderTH      string            `json:"sub_header_th"`
	SubHeaderEN      string            `json:"sub_header_en"`
	StartDate        time.Time         `json:"start_date"`
	EndDate          time.Time         `json:"end_date"`
	ProductListRefID string            `json:"product_list_ref_id"`
	Components       []componentRecord `json:"components"`
	CreatedAt        time.Time         `json:"created_at"`
	CreatedBy        string            `json:"created_by"`
	UpdatedAt        time.Time         `json:"updated_at"`
	UpdatedBy        string            `json:"updated_by"`
}

type componentRecord struct {
	Type     string `json:"type"`
	SubType  string `json:"sub_type"`
	ImageURL string `json:"image_url"`
	IsActive bool   `json:"is_active"`
	DataSrc  string `json:"data_src"`
}

func toRecord(w domain.Widget) widgetRecord {
	cs := make([]componentRecord, len(w.Components))
	for i, c := range w.Components {
		cs[i] = componentRecord{Type: string(c.Type), SubType: c.SubType, ImageURL: c.ImageURL, IsActive: c.IsActive, DataSrc: c.DataSrc}
	}
	return widgetRecord{
		ID: w.ID, HeaderTH: w.HeaderTH, HeaderEN: w.HeaderEN, SubHeaderTH: w.SubHeaderTH, SubHeaderEN: w.SubHeaderEN,
		StartDate: w.StartDate, EndDate: w.EndDate, ProductListRefID: w.ProductListRefID, Components: cs,
		CreatedAt: w.CreatedAt, CreatedBy: w.CreatedBy, UpdatedAt: w.UpdatedAt, UpdatedBy: w.UpdatedBy,
	}
}

func (r widgetRecord) toDomain() domain.Widget {
	cs := make([]domain.Component, len(r.Components))
	for i, c := range r.Components {
		cs[i] = domain.Component{Type: domain.ComponentType(c.Type), SubType: c.SubType, ImageURL: c.ImageURL, IsActive: c.IsActive, DataSrc: c.DataSrc}
	}
	return domain.Widget{
		ID: r.ID, HeaderTH: r.HeaderTH, HeaderEN: r.HeaderEN, SubHeaderTH: r.SubHeaderTH, SubHeaderEN: r.SubHeaderEN,
		StartDate: r.StartDate, EndDate: r.EndDate, ProductListRefID: r.ProductListRefID, Components: cs,
		CreatedAt: r.CreatedAt, CreatedBy: r.CreatedBy, UpdatedAt: r.UpdatedAt, UpdatedBy: r.UpdatedBy,
	}
}
