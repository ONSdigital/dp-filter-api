package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveDuplicateEvents(t *testing.T) {
	Convey("Given the stored filter output doesn't contain any events", t, func() {
		current := &Filter{
			ID:     "1234",
			Events: []*Event{},
		}

		Convey("When filter a new filter provides no events", func() {
			newFilter := &Filter{
				ID:     "1234",
				Events: []*Event{},
			}

			Convey("Then the result should still be empty", func() {
				newFilter.RemoveDuplicateEvents(current)
				So(len(newFilter.Events), ShouldEqual, 0)
			})
		})

		Convey("When filter a new filter provides events", func() {
			newFilter := &Filter{
				Events: []*Event{
					{
						Type: "Event1",
						Time: time.Now(),
					},
				},
			}

			Convey("Then all of them should be kept", func() {
				newFilter.RemoveDuplicateEvents(current)
				So(len(newFilter.Events), ShouldEqual, 1)
			})
		})
	})

	Convey("Given the stored filter output does contain events", t, func() {
		e1 := &Event{
			Type: "Event1",
			Time: time.Now(),
		}

		e2 := &Event{
			Type: "Event2",
			Time: time.Now(),
		}

		current := &Filter{
			Events: []*Event{e1, e2},
		}

		Convey("When filter a new filter provides no events", func() {
			newFilter := &Filter{
				ID:     "1234",
				Events: []*Event{},
			}

			Convey("Then the result should be empty", func() {
				newFilter.RemoveDuplicateEvents(current)
				So(len(newFilter.Events), ShouldEqual, 0)
			})
		})

		Convey("When filter a new filter provides a new event", func() {
			newE := &Event{
				Type: "New Event",
				Time: time.Now(),
			}
			newFilter := &Filter{
				Events: []*Event{newE},
			}

			Convey("Then the new event should be kept", func() {
				newFilter.RemoveDuplicateEvents(current)
				So(len(newFilter.Events), ShouldEqual, 1)
				So(newFilter.Events[0], ShouldEqual, newE)
			})
		})

		Convey("When filter a new filter provides an existing event", func() {
			newFilter := &Filter{
				Events: []*Event{e1},
			}

			Convey("Then the result should be empty", func() {
				newFilter.RemoveDuplicateEvents(current)
				So(len(newFilter.Events), ShouldEqual, 0)
			})
		})
	})
}

func TestValidate(t *testing.T) {
	Convey("Given an event without a time field", t, func() {
		event := &Event{Type: EventFilterOutputCreated}

		Convey("When Validate is called", func() {
			err := event.Validate()
			Convey("Then the error should contain only the time field", func() {
				So(err, ShouldNotBeEmpty)
				So(err.Error(), ShouldContainSubstring, "missing mandatory fields")
				So(err.Error(), ShouldNotContainSubstring, "event.type")
				So(err.Error(), ShouldContainSubstring, "event.time")
			})
		})
	})

	Convey("Given an event without a type field", t, func() {
		event := &Event{Time: time.Now()}

		Convey("When Validate is called", func() {
			err := event.Validate()
			Convey("Then the error should contain only the time field", func() {
				So(err, ShouldNotBeEmpty)
				So(err.Error(), ShouldContainSubstring, "missing mandatory fields")
				So(err.Error(), ShouldContainSubstring, "event.type")
				So(err.Error(), ShouldNotContainSubstring, "event.time")
			})
		})
	})

	Convey("Given an event with both fields populated", t, func() {
		event := &Event{
			Time: time.Now(),
			Type: EventFilterOutputCreated,
		}

		Convey("When Validate is called", func() {
			err := event.Validate()
			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an event with invalid type info", t, func() {
		event := &Event{
			Time: time.Now(),
			Type: "anewevent",
		}

		Convey("When Validate is called", func() {
			err := event.Validate()
			Convey("Then no error should be returned", func() {
				So(err, ShouldNotBeEmpty)
				So(err.Error(), ShouldContainSubstring, "invalid event type")
				So(err.Error(), ShouldContainSubstring, "anewevent")
			})
		})
	})
}
