package domain

import "time"

// EventType names a widget change.
type EventType string

const (
	EventWidgetCreated EventType = "widget.created"
	EventWidgetUpdated EventType = "widget.updated"
	EventWidgetDeleted EventType = "widget.deleted"
)

// WidgetEvent is a change to a widget that the core publishes.
type WidgetEvent struct {
	Type     EventType
	WidgetID string
	Actor    string
	At       time.Time
}
