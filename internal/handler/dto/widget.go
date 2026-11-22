// Package dto defines the HTTP wire shapes and their mapping to domain types.
package dto

import (
	"fmt"
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// DateLayout is the wire date format, DD-MM-YYYY.
const DateLayout = "02-01-2006"

// ComponentInput is a component in a request.
type ComponentInput struct {
	Type     string `json:"type"`
	SubType  string `json:"sub_type,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	IsActive bool   `json:"is_active"`
}

// Component is a component in a response.
type Component struct {
	Type     string `json:"type"`
	SubType  string `json:"sub_type,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	IsActive bool   `json:"is_active"`
	DataSrc  string `json:"data_src,omitempty"`
}

// WidgetBody is the editable payload shared by create and update.
type WidgetBody struct {
	HeaderNameTH     string           `json:"header_name_th"`
	HeaderNameEN     string           `json:"header_name_en"`
	SubHeaderNameTH  string           `json:"sub_header_name_th"`
	SubHeaderNameEN  string           `json:"sub_header_name_en"`
	StartDate        string           `json:"start_date" example:"21-12-2025"`
	EndDate          string           `json:"end_date" example:"31-12-2025"`
	ProductListRefID string           `json:"product_list_ref_id"`
	Components       []ComponentInput `json:"widgets"`
}

// ToDomain parses the wire dates in loc.
func (b WidgetBody) ToDomain(loc *time.Location) (domain.WidgetInput, error) {
	start, err := parseDay(b.StartDate, loc, 0, 0, 0)
	if err != nil {
		return domain.WidgetInput{}, fmt.Errorf("%w: start_date: %v", domain.ErrInvalidInput, err)
	}
	end, err := parseDay(b.EndDate, loc, 23, 59, 59)
	if err != nil {
		return domain.WidgetInput{}, fmt.Errorf("%w: end_date: %v", domain.ErrInvalidInput, err)
	}

	cs := make([]domain.Component, len(b.Components))
	for i, c := range b.Components {
		cs[i] = domain.Component{
			Type:     domain.ComponentType(c.Type),
			SubType:  c.SubType,
			ImageURL: c.ImageURL,
			IsActive: c.IsActive,
		}
	}

	return domain.WidgetInput{
		HeaderTH:         b.HeaderNameTH,
		HeaderEN:         b.HeaderNameEN,
		SubHeaderTH:      b.SubHeaderNameTH,
		SubHeaderEN:      b.SubHeaderNameEN,
		StartDate:        start,
		EndDate:          end,
		ProductListRefID: b.ProductListRefID,
		Components:       cs,
	}, nil
}

func parseDay(s string, loc *time.Location, h, m, sec int) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	d, err := time.ParseInLocation(DateLayout, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("want %s, got %q", DateLayout, s)
	}
	return time.Date(d.Year(), d.Month(), d.Day(), h, m, sec, 0, loc), nil
}

func formatDay(t time.Time, loc *time.Location) string { return t.In(loc).Format(DateLayout) }

func componentsFromDomain(cs []domain.Component) []Component {
	out := make([]Component, len(cs))
	for i, c := range cs {
		out[i] = Component{Type: string(c.Type), SubType: c.SubType, ImageURL: c.ImageURL, IsActive: c.IsActive, DataSrc: c.DataSrc}
	}
	return out
}

// CreateWidgetRequest is the POST /widgets body.
type CreateWidgetRequest struct {
	WidgetBody
}

// CreateWidgetResponse returns the new widget ID.
type CreateWidgetResponse struct {
	WidgetID string `json:"widget_id"`
}

// UpdateWidgetRequest is the PUT /widgets/:widget-id path and body.
type UpdateWidgetRequest struct {
	WidgetID string `uri:"widget-id" json:"-"`
	WidgetBody
}

// WidgetIDRequest is the :widget-id path parameter.
type WidgetIDRequest struct {
	WidgetID string `uri:"widget-id"`
}

// WidgetDetailResponse is one widget with its status.
type WidgetDetailResponse struct {
	WidgetID         string      `json:"widget_id"`
	HeaderNameTH     string      `json:"header_name_th"`
	HeaderNameEN     string      `json:"header_name_en"`
	SubHeaderNameTH  string      `json:"sub_header_name_th"`
	SubHeaderNameEN  string      `json:"sub_header_name_en"`
	StartDate        string      `json:"start_date"`
	EndDate          string      `json:"end_date"`
	ProductListRefID string      `json:"product_list_ref_id"`
	Components       []Component `json:"widgets"`
	Status           string      `json:"status"`
}

// WidgetDetailFromDomain formats the dates in loc.
func WidgetDetailFromDomain(v domain.WidgetView, loc *time.Location) WidgetDetailResponse {
	return WidgetDetailResponse{
		WidgetID:         v.ID,
		HeaderNameTH:     v.HeaderTH,
		HeaderNameEN:     v.HeaderEN,
		SubHeaderNameTH:  v.SubHeaderTH,
		SubHeaderNameEN:  v.SubHeaderEN,
		StartDate:        formatDay(v.StartDate, loc),
		EndDate:          formatDay(v.EndDate, loc),
		ProductListRefID: v.ProductListRefID,
		Components:       componentsFromDomain(v.Components),
		Status:           string(v.Status),
	}
}

// ListWidgetsRequest is the GET /widgets paging query.
type ListWidgetsRequest struct {
	PageID   int `form:"page_id"`
	PageSize int `form:"page_size"`
}

// ToDomain maps the query to a domain.Page.
func (r ListWidgetsRequest) ToDomain() domain.Page {
	return domain.Page{Number: r.PageID, Size: r.PageSize}
}

// WidgetSummary is one row of a widget list.
type WidgetSummary struct {
	No           int    `json:"no"`
	WidgetID     string `json:"widget_id"`
	HeaderNameTH string `json:"header_name_th"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       string `json:"status"`
}

// Pagination describes the returned page.
type Pagination struct {
	PageID        int `json:"page_id"`
	PageSize      int `json:"page_size"`
	LastPage      int `json:"last_page"`
	TotalElements int `json:"total_elements"`
}

// ListWidgetsResponse is one page of widgets.
type ListWidgetsResponse struct {
	Items      []WidgetSummary `json:"widgets"`
	Pagination Pagination      `json:"pagination"`
}

// ListWidgetsFromDomain formats the dates in loc.
func ListWidgetsFromDomain(r domain.PageResult[domain.WidgetView], loc *time.Location) ListWidgetsResponse {
	items := make([]WidgetSummary, len(r.Items))
	for i, v := range r.Items {
		items[i] = WidgetSummary{
			No:           r.Page.Offset() + i + 1,
			WidgetID:     v.ID,
			HeaderNameTH: v.HeaderTH,
			StartDate:    formatDay(v.StartDate, loc),
			EndDate:      formatDay(v.EndDate, loc),
			Status:       string(v.Status),
		}
	}
	return ListWidgetsResponse{
		Items: items,
		Pagination: Pagination{
			PageID:        r.Page.Number,
			PageSize:      r.Page.Size,
			LastPage:      r.LastPage(),
			TotalElements: r.Total,
		},
	}
}
