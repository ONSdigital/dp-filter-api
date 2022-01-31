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

	updatedEvents := filter.Events
	for i, event := range updatedEvents {
		for _, ce := range currentFilter.Events {
			if event == ce {
				//remove duplicate events from the new filter array
				updatedEvents = removeEvents(updatedEvents, i)
			}
		}
	}

	filter.Events = updatedEvents
}

//remove event at index from the list or return the full list if the index is out of bounds
func removeEvents(list []*Event, i int) []*Event {
	if i >= len(list) {
		return list
	}

	if i == 0 {
		return list[1:]
	}

	return append(list[:i], list[i+1:]...)

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
