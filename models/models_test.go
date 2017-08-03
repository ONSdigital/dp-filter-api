package models

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// A Mock io.reader to trigger errors on reading
type reader struct {
}

func (f reader) Read(bytes []byte) (int, error) {
	return 0, fmt.Errorf("Reader failed")
}

func TestCreateFilterJobWithNoBody(t *testing.T) {
	Convey("When a filter job message has no body, an error is returned", t, func() {
		_, err := CreateFilter(reader{})
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Failed to read message body"))
	})
}

func TestCreateFilterJobWithInvalidJson(t *testing.T) {
	Convey("When a filter job message is missing dataset field, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{\"version\":\"1\",\"edition\":\"1\"}"))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterCreation()
		missingFields := []string{"dataset"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter job message is missing version field, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{\"dataset\":\"1\",\"edition\":\"1\"}"))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterCreation()
		missingFields := []string{"version"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter job message is missing edition field, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{\"dataset\":\"1\",\"version\":\"1\"}"))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterCreation()
		missingFields := []string{"edition"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter job message has an empty json body, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{ }"))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterCreation()
		missingFields := []string{"dataset", "edition", "version"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})
}

func TestCreateJobWithInvalidJson(t *testing.T) {
	Convey("When a job message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader("{ ")
		_, err := CreateFilter(reader)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Failed to parse json body"))
	})
}

func TestCreateFilterJobWithValidJSON(t *testing.T) {
	Convey("When a filter job has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader("{\"dataset\":\"1\",\"version\":\"1\",\"edition\":\"1\"}")
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.ValidateFilterCreation(), ShouldBeNil)
		So(filter.FilterID, ShouldNotBeNil)
		So(filter.DataSet, ShouldEqual, "1")
		So(filter.Version, ShouldEqual, "1")
		So(filter.Edition, ShouldEqual, "1")
		So(filter.State, ShouldEqual, "created")
	})
}
