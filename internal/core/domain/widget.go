package domain

import (
	"fmt"
	"time"
)

// Status is a widget's schedule state at a point in time; see Widget.StatusAt.
type Status string

const (
	StatusNew      Status = "NEW"
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

// ComponentType is the kind of block a Component renders.
type ComponentType string

const (
	ComponentInformation ComponentType = "INFORMATION"
	ComponentProduct     ComponentType = "PRODUCT"
	ComponentCategory    ComponentType = "CATEGORY"
)

// Component is one block rendered inside a widget.
type Component struct {
	Type     ComponentType
	SubType  string
	ImageURL string
	IsActive bool
	DataSrc  string
}

// Widget is the widget aggregate. Build it with NewWidget and change it with Revise.
type Widget struct {
	ID               string
	HeaderTH         string
	HeaderEN         string
	SubHeaderTH      string
	SubHeaderEN      string
	StartDate        time.Time
	EndDate          time.Time
	ProductListRefID string
	Components       []Component
	CreatedAt        time.Time
	CreatedBy        string
	UpdatedAt        time.Time
	UpdatedBy        string
}

// WidgetInput is the caller-editable part of a Widget plus the acting user.
type WidgetInput struct {
	Actor            string
	HeaderTH         string
	HeaderEN         string
	SubHeaderTH      string
	SubHeaderEN      string
	StartDate        time.Time
	EndDate          time.Time
	ProductListRefID string
	Components       []Component
}

// Validate returns ErrInvalidInput for a missing field, reversed dates or an unknown component type.
func (in WidgetInput) Validate() error {
	switch {
	case in.Actor == "":
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	case in.HeaderTH == "":
		return fmt.Errorf("%w: header_name_th is required", ErrInvalidInput)
	case in.StartDate.IsZero() || in.EndDate.IsZero():
		return fmt.Errorf("%w: start_date and end_date are required", ErrInvalidInput)
	case in.EndDate.Before(in.StartDate):
		return fmt.Errorf("%w: end_date is before start_date", ErrInvalidInput)
	}
	for i, c := range in.Components {
		switch c.Type {
		case ComponentInformation, ComponentProduct, ComponentCategory:
		default:
			return fmt.Errorf("%w: widgets[%d].type %q is unknown", ErrInvalidInput, i, c.Type)
		}
	}
	return nil
}

// NewWidget builds a valid Widget or returns ErrInvalidInput.
func NewWidget(id string, in WidgetInput, now time.Time) (Widget, error) {
	if err := in.Validate(); err != nil {
		return Widget{}, err
	}
	w := Widget{ID: id, CreatedAt: now, CreatedBy: in.Actor}
	w.apply(in, now)
	return w, nil
}

// Revise returns a copy of w with the editable fields from in; ID and Created* are kept.
func (w Widget) Revise(in WidgetInput, now time.Time) (Widget, error) {
	if err := in.Validate(); err != nil {
		return Widget{}, err
	}
	w.apply(in, now)
	return w, nil
}

func (w *Widget) apply(in WidgetInput, now time.Time) {
	w.HeaderTH, w.HeaderEN = in.HeaderTH, in.HeaderEN
	w.SubHeaderTH, w.SubHeaderEN = in.SubHeaderTH, in.SubHeaderEN
	w.StartDate, w.EndDate = in.StartDate, in.EndDate
	w.ProductListRefID = in.ProductListRefID
	w.Components = append([]Component(nil), in.Components...)
	w.UpdatedAt, w.UpdatedBy = now, in.Actor
}

// StatusAt returns w's Status at now.
func (w Widget) StatusAt(now time.Time) Status {
	switch {
	case now.Before(w.StartDate):
		return StatusNew
	case now.After(w.EndDate):
		return StatusInactive
	default:
		return StatusActive
	}
}

// WidgetView is the read model returned by the service: the aggregate plus
// values derived at read time.
type WidgetView struct {
	Widget
	Status Status
}
