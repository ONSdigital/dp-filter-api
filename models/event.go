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

// Event captures the time certain stages of filter response were completed
type Event struct {
	Type string    `bson:"type,omitempty" json:"type"`
	Time time.Time `bson:"time,omitempty" json:"time"`
}

func (filter *Filter) RemoveDuplicateEvents(currentFilter *Filter) {
	if len(filter.Events) == 0 || len(currentFilter.Events) == 0 {
		return
	}

	events := []*Event{}

	for _, e := range filter.Events {
		found := false
		for _, ce := range currentFilter.Events {
			// compare the values not the pointers
			if *e == *ce {
				found = true
				break
			}
		}

		if !found {
			events = append(events, e)
		}
	}

	filter.Events = events
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
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	switch e.Type {
	case EventFilterOutputCreated,
		EventFilterOutputQueryStart, EventFilterOutputQueryEnd,
		EventFilterOutputCSVGenStart, EventFilterOutputCSVGenEnd,
		EventFilterOutputXLSXGenStart, EventFilterOutputXLSXGenEnd,
		EventFilterOutputCompleted:
		break
	default:
		return fmt.Errorf("invalid event type provided: %v", e.Type)
	}

	return nil
}
