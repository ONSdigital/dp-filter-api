package models

import (
	"fmt"
	"time"
)

// A list of event types
const (
	EventFilterOutputCreated      = "FilterOutputCreated"
	EventFilterOutputQueryStart   = "FilterOutputQueryStart"
	EventFilterOutputQueryEnd     = "FilterOutputQueryEnd"
	EventFilterOutputCSVGenStart  = "FilterOutputCSVGenStart"
	EventFilterOutputCSVGenEnd    = "FilterOutputCSVGenEnd"
	EventFilterOutputXLSXGenStart = "FilterOutputXLSXGenStart"
	EventFilterOutputXLSXGenEnd   = "FilterOutputXLSXGenEnd"
	EventFilterOutputCompleted    = "FilterOutputCompleted"
)

type Event struct {
	Type string    `bson:"type,omitempty" json:"type"`
	Time time.Time `bson:"time,omitempty" json:"time"`
}

// Validate the content of the event structure are
// recognized event names and a non-zero time
func (e *Event) Validate() error {
	var missingFields []string

	if e.Time.IsZero() {
		missingFields = append(missingFields, "event.time")

	}

	if e.Type == "" {
		missingFields = append(missingFields, "event.type")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	switch e.Type {
	case EventFilterOutputCreated,
		EventFilterOutputQueryStart, EventFilterOutputQueryEnd,
		EventFilterOutputCSVGenStart, EventFilterOutputCSVGenEnd,
		EventFilterOutputXLSXGenStart, EventFilterOutputXLSXGenEnd,
		EventFilterOutputCompleted:
		break
	default:
		return fmt.Errorf("Invalid event type provided: %v", e.Type)

	}

	return nil
}
