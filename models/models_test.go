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
	return 0, fmt.Errorf("reader failed")
}

func TestCreateFilterBlueprintWithValidJSON(t *testing.T) {
	Convey("When a filter blueprint has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		filter, err := CreateNewFilter(reader)
		So(err, ShouldBeNil)
		So(filter.ValidateNewFilter(), ShouldBeNil)
		So(filter.Dataset, ShouldNotBeNil)
		So(filter.Dataset.ID, ShouldEqual, "1")
		So(filter.Dataset.Edition, ShouldEqual, "1")
		So(filter.Dataset.Version, ShouldEqual, 1)
	})
}

func TestCreateFilterWithNoBody(t *testing.T) {
	Convey("When a filter message has no body, an error is returned", t, func() {
		_, err := CreateNewFilter(reader{})
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorReadingBody)
	})

	Convey("When a filter message has an empty body, an error is returned", t, func() {
		filter, err := CreateNewFilter(strings.NewReader("{}"))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorNoData)
		So(filter, ShouldNotBeNil)
	})
}

func TestCreateFilterBlueprintWithInvalidJson(t *testing.T) {
	Convey("When a filter blueprint message is missing dataset fields, an error is returned", t, func() {
		filter, err := CreateNewFilter(strings.NewReader(`{"dataset":{"version":1} }`))
		So(err, ShouldBeNil)

		err = filter.ValidateNewFilter()
		missingFields := []string{"dataset.edition", "dataset.id"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("missing mandatory fields: %v", missingFields))
	})

	Convey("When a filter blueprint message has an empty json body, an error is returned", t, func() {
		filter, err := CreateNewFilter(strings.NewReader("{ }"))
		So(err, ShouldBeNil)

		err = filter.ValidateNewFilter()
		missingFields := []string{"dataset.version", "dataset.edition", "dataset.id"}
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("missing mandatory fields: %v", missingFields))
	})
}

func TestCreateBlueprintWithInvalidJson(t *testing.T) {
	Convey("When a job message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader(`{`)
		_, err := CreateFilter(reader)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorParsingBody)
	})
}

func TestValidateFilterOutputUpdate(t *testing.T) {
	Convey("Given the filterOutput doesn't contain any forbidden fields", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"href":"some-test-url","size":"12mb","private":"some-private-link"}}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		currentFilter := &Filter{Published: &Unpublished}

		Convey("When filter is validated then no errors are returned", func() {
			err = filter.ValidateFilterOutputUpdate(currentFilter)
			So(err, ShouldBeNil)
		})
	})

	Convey("Given the filterOutput contains dimensions", t, func() {
		reader := strings.NewReader(`{"dimensions":[{"dimension_url":"some-test-dimension-url","name":"age","options":["24"]}]}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		currentFilter := &Filter{Published: &Unpublished}

		Convey("When filter is validated then an error is returned", func() {
			err = filter.ValidateFilterOutputUpdate(currentFilter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [dimensions]"))
		})
	})

	Convey("Given the filterOutput contains instance id", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		currentFilter := &Filter{Published: &Unpublished}

		Convey("When filter is validated then an error is returned", func() {
			err = filter.ValidateFilterOutputUpdate(currentFilter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [instance_id]"))
		})
	})

	Convey("Given the filterOutput contains filter_id", t, func() {
		reader := strings.NewReader(`{"filter_id":"87654321"}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		currentFilter := &Filter{Published: &Unpublished}

		Convey("When filter is validated then an error is returned", func() {
			err = filter.ValidateFilterOutputUpdate(currentFilter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [filter_id]"))
		})
	})

	Convey("Given the filterOutput contains filter_id, instance_id and dimensions", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678","filter_id":"87654321","dimensions":[{"dimension_url":"some-test-dimension-url","name":"age","options":["24"]}]}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)

		currentFilter := &Filter{Published: &Unpublished}

		Convey("When filter is validated then an error is returned", func() {
			err = filter.ValidateFilterOutputUpdate(currentFilter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [instance_id dimensions filter_id]"))
		})
	})
}

func TestValidateFilterOutputDownloadsUpdate(t *testing.T) {
	Convey("Given the version is published and the current filter has private"+
		"downloads for csv and xls but not public ones", t, func() {
		downloads := &Downloads{
			CSV: &DownloadItem{
				HRef:    "/filter-outputs/87654321.csv",
				Private: "private-link",
				Size:    "12mb",
			},
			XLS: &DownloadItem{
				HRef:    "/filter-outputs/87654321.xls",
				Private: "private-link",
				Size:    "24mb",
			},
		}

		currentFilter := &Filter{Published: &Published, Downloads: downloads}

		Convey("When filter update contains csv private link", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"href":"some-test-url","size":"12mb","private":"some-private-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.csv.private]"))
			})
		})

		Convey("When filter update contains xls private link", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"some-test-url","size":"12mb","private":"some-private-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.xls.private]"))
			})
		})

		Convey("When filter update contains csv and xls private link", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"some-test-url","size":"12mb","private":"some-private-link"}, "csv":{"href":"some-test-url","size":"12mb","private":"some-private-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.csv.private downloads.xls.private]"))
			})
		})

		Convey("When filter update contains a csv public link but not a private link", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"href":"some-test-url","size":"12mb","public":"some-public-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation NO error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldBeNil)
			})
		})

		Convey("When filter update contains a xls public link but not a private link", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"some-test-url","size":"12mb","public":"some-public-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation NO error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given the version is published and the current filter has public downloads link for csv and xls", t, func() {
		downloads := &Downloads{
			CSV: &DownloadItem{
				HRef:    "/filter-outputs/87654321.csv",
				Private: "private-link",
				Public:  "public-link",
				Size:    "12mb",
			},
			XLS: &DownloadItem{
				HRef:    "/filter-outputs/87654321.xls",
				Private: "private-link",
				Public:  "public-link",
				Size:    "24mb",
			},
		}
		currentFilter := &Filter{Published: &Published, Downloads: downloads}

		Convey("When filter update contains csv public link", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"href":"some-test-url","size":"12mb","public":"some-public-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.csv]"))
			})
		})

		Convey("When filter update contains xls public link", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"some-test-url","size":"12mb","public":"some-public-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.xls]"))
			})
		})

		Convey("When filter update contains csv and xls public link", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"href":"some-test-url","size":"12mb","public":"some-public-link"},"xls":{"href":"some-test-url","size":"12mb","public":"some-public-link"}}}`)
			filter, err := CreateFilter(reader)
			So(err, ShouldBeNil)

			Convey("Then on validation an error is returned", func() {
				err = filter.ValidateFilterOutputUpdate(currentFilter)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [downloads.csv downloads.xls]"))
			})
		})
	})
}

func TestValidateFilterBlueprintUpdate(t *testing.T) {
	Convey("Given the filter blueprint update doesn't contain any forbidden fields", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.Dataset.Version, ShouldEqual, 2)

		Convey("When filter is validated then no errors are returned", func() {
			err = ValidateFilterBlueprintUpdate(filter)
			So(err, ShouldBeNil)
		})
	})

	Convey("Given the filter blueprint update contains dataset id", t, func() {
		reader := strings.NewReader(`{"dataset":{"id":"123"}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.Dataset.ID, ShouldEqual, "123")

		Convey("When filter is validated then an error is returned", func() {
			err = ValidateFilterBlueprintUpdate(filter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [dataset.id]"))
		})
	})

	Convey("Given the filter blueprint update contains dataset edition", t, func() {
		reader := strings.NewReader(`{"dataset":{"edition":"2018"}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.Dataset.Edition, ShouldEqual, "2018")

		Convey("When filter is validated then an error is returned", func() {
			err = ValidateFilterBlueprintUpdate(filter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [dataset.edition]"))
		})
	})

	Convey("Given the filter blueprint update contains both the dataset id and edition", t, func() {
		reader := strings.NewReader(`{"dataset":{"id":"123", "edition":"2018"}}`)
		filter, err := CreateFilter(reader)
		So(err, ShouldBeNil)
		So(filter.Dataset.ID, ShouldEqual, "123")
		So(filter.Dataset.Edition, ShouldEqual, "2018")

		Convey("When filter is validated then an error is returned", func() {
			err = ValidateFilterBlueprintUpdate(filter)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("forbidden from updating the following fields: [dataset.id dataset.edition]"))
		})
	})
}

func TestFilterHash(t *testing.T) {
	testFilter := func() Filter {
		return Filter{
			FilterID:   "myFilter",
			InstanceID: "myInstance",
			State:      CreatedState,
			Dimensions: []Dimension{
				{
					URL:     "http://dimensions.co.uk/dim1",
					Name:    "dim1",
					Options: []string{"op11", "op12", "op13"},
				},
				{
					URL:     "http://dimensions.co.uk/dim2",
					Name:    "dim2",
					Options: []string{"op21"},
				},
			},
		}
	}

	Convey("Given a filter blueprint with some data", t, func() {
		filter := testFilter()

		Convey("We can generate a valid hash", func() {
			h, err := filter.Hash(nil)
			So(err, ShouldBeNil)
			So(len(h), ShouldEqual, 40)

			Convey("Then hashing it twice, produces the same result", func() {
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldEqual, h)
			})

			Convey("Then storing the hash as its ETag value and hashing it again, produces the same result (field is ignored) and ETag field is preserved", func() {
				filter.ETag = h
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldEqual, h)
				So(filter.ETag, ShouldEqual, h)
			})

			Convey("Then another filter with exactly the same data will resolve to the same hash", func() {
				filter2 := testFilter()
				hash, err := filter2.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldEqual, h)
			})

			Convey("Then if a filter value is modified, its hash changes", func() {
				filter.State = CompletedState
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldNotEqual, h)
			})

			Convey("Then if a dimension is added to the filter, its hash changes", func() {
				filter.Dimensions = append(filter.Dimensions, Dimension{Name: "dim3"})
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldNotEqual, h)
			})

			Convey("Then if a dimension is removed from the filter, its hash changes", func() {
				filter.Dimensions = []Dimension{
					{
						URL:     "http://dimensions.co.uk/dim1",
						Name:    "dim1",
						Options: []string{"op11", "op12", "op13"},
					},
				}
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldNotEqual, h)
			})

			Convey("Then if a dimension option is added to a filter dimension, its hash changes", func() {
				filter.Dimensions[0].Options = append(filter.Dimensions[0].Options, "op14")
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldNotEqual, h)
			})

			Convey("Then if a dimension option is removed from a filter dimension, its hash changes", func() {
				filter.Dimensions[0].Options = []string{"op11", "op13"}
				hash, err := filter.Hash(nil)
				So(err, ShouldBeNil)
				So(hash, ShouldNotEqual, h)
			})
		})
	})
}
