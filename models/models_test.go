package models

import (
	"errors"
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

func TestCreateFilterBlueprintWithValidJSON(t *testing.T) {
	Convey("When a filter blueprint has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.ValidateFilterBlueprint(), ShouldBeNil)
		So(filter.FilterID, ShouldNotBeNil)
		So(filter.InstanceID, ShouldEqual, "12345678")
		So(filter.State, ShouldEqual, "")
	})

	Convey("When a filter blueprint has a valid json body with a state, a message is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678","state":"created"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.ValidateFilterBlueprint(), ShouldBeNil)
		So(filter.FilterID, ShouldNotBeNil)
		So(filter.InstanceID, ShouldEqual, "12345678")
		So(filter.State, ShouldEqual, "")
	})
}

func TestCreateFilterWithNoBody(t *testing.T) {
	Convey("When a filter message has no body, an error is returned", t, func() {
		_, err := CreateFilter(reader{})
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Failed to read message body"))
	})

	Convey("When a filter message has an empty body, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{}"))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Bad request - Missing data in body"))
		So(filter, ShouldNotBeNil)
	})
}

func TestCreateFilterBlueprintWithInvalidJson(t *testing.T) {
	Convey("When a filter blueprint message is missing instance_id field, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader(`{"state":"created"}`))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterBlueprint()
		missingFields := []string{"instance_id"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter blueprint message has an empty json body, an error is returned", t, func() {
		filter, err := CreateFilter(strings.NewReader("{ }"))
		So(err, ShouldBeNil)

		err = filter.ValidateFilterBlueprint()
		missingFields := []string{"instance_id"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Missing mandatory fields: %v", missingFields))
	})
}

func TestCreateBlueprintWithInvalidJson(t *testing.T) {
	Convey("When a job message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader(`{`)
		_, err := CreateFilter(reader)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Failed to parse json body"))
	})
}

func TestValidateFilterOutputUpdate(t *testing.T) {
	Convey("Given the filterOutput doesn't contain any forbidden fields", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"some-test-url","size":"12mb"}}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		Convey("When filter is validated then no errors are returned", func() {

			err = filter.ValidateFilterOutputUpdate()
			So(err, ShouldBeNil)
		})
	})

	Convey("Given the filterOutput contains dimensions", t, func() {
		reader := strings.NewReader(`{"dimensions":[{"dimension_url":"some-test-dimension-url","name":"age","options":["24"]}]}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		Convey("When filter is validated then an error is returned", func() {

			err = filter.ValidateFilterOutputUpdate()
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Forbidden from updating the following fields: [dimensions]"))
		})
	})

	Convey("Given the filterOutput contains instance id", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		Convey("When filter is validated then an error is returned", func() {

			err = filter.ValidateFilterOutputUpdate()
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Forbidden from updating the following fields: [instance_id]"))
		})
	})

	Convey("Given the filterOutput contains filter_id", t, func() {
		reader := strings.NewReader(`{"filter_id":"87654321"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		Convey("When filter is validated then an error is returned", func() {

			err = filter.ValidateFilterOutputUpdate()
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Forbidden from updating the following fields: [filter_id]"))
		})
	})

	Convey("Given the filterOutput contains filter_id, instance_id and dimensions", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678","filter_id":"87654321","dimensions":[{"dimension_url":"some-test-dimension-url","name":"age","options":["24"]}]}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		Convey("When filter is validated then an error is returned", func() {

			err = filter.ValidateFilterOutputUpdate()
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Forbidden from updating the following fields: [instance_id dimensions filter_id]"))
		})
	})
}
