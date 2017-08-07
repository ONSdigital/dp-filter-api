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
	Convey("When a filter job message is missing dataset_filter_id field, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{\"state\":\"created\"}"))
		So(err, ShouldBeNil)

		err = filter.Validate()
		missingFields := []string{"dataset_filter_id"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter job message has an empty json body, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{ }"))
		So(err, ShouldBeNil)

		err = filter.Validate()
		missingFields := []string{"dataset_filter_id"}
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
		reader := strings.NewReader("{\"dataset_filter_id\":\"12345678\"}")
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.Validate(), ShouldBeNil)
		So(filter.FilterID, ShouldNotBeNil)
		So(filter.DataSetFilterID, ShouldEqual, "12345678")
		So(filter.State, ShouldEqual, "created")
	})
}
