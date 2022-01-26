package mongo

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testFilterID      = "testFilterID"
	testInstanceID    = "testInstanceID"
	testDimensionName = "testDimensionName"
)

func testFilter() *models.Filter {
	notPublished := false
	f := &models.Filter{
		InstanceID: testInstanceID,
		FilterID:   testFilterID,
		State:      models.CreatedState,
		Published:  &notPublished,
		Dimensions: []models.Dimension{
			{
				Name:    testDimensionName,
				Options: []string{"op1", "op2", "op3"},
			},
		},
	}
	eTag0, err := f.Hash(nil)
	So(err, ShouldBeNil)
	f.ETag = eTag0
	return f
}

func TestGetNewETagForUpdate(t *testing.T) {

	Convey("Given a filer that we want to update", t, func() {

		currentFilter := testFilter()

		published := true
		update := &models.Filter{
			InstanceID: "testInstanceIDUpdated",
			Published:  &published,
		}

		Convey("getNewETagForUpdate returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := newETagForUpdate(currentFilter, update)
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)

			Convey("Applying the same update to a different filter results in a different ETag", func() {
				filter2 := testFilter()
				filter2.FilterID = "otherFilter"
				eTag2, err := newETagForUpdate(filter2, update)
				So(err, ShouldBeNil)
				So(eTag2, ShouldNotEqual, eTag1)
			})

			Convey("Applying a different update to the same filter results in a different ETag", func() {
				update2 := &models.Filter{
					InstanceID: "anotherInstanceID",
					Published:  &published,
				}
				eTag3, err := newETagForUpdate(currentFilter, update2)
				So(err, ShouldBeNil)
				So(eTag3, ShouldNotEqual, eTag1)
			})
		})
	})
}

func TestGetNewETagForDimensionOperations(t *testing.T) {

	Convey("Given a filer whose dimensions we want to update", t, func() {

		currentFilter := testFilter()

		dims := []models.Dimension{
			{URL: "url1", Name: "dim1"},
		}

		Convey("getNewETagForAddDimensions returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := newETagForAddDimensions(currentFilter, testFilterID, dims)
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)

			Convey("Applying the same update to a different filter results in a different ETag", func() {
				filter2 := testFilter()
				filter2.FilterID = "otherFilter"
				eTag2, err := newETagForAddDimensions(filter2, testFilterID, dims)
				So(err, ShouldBeNil)
				So(eTag2, ShouldNotEqual, eTag1)
			})

			Convey("Applying a different update to the same filter results in a different ETag", func() {
				dims2 := []models.Dimension{
					{URL: "url2", Name: "dim2"},
				}
				eTag3, err := newETagForAddDimensions(currentFilter, testFilterID, dims2)
				So(err, ShouldBeNil)
				So(eTag3, ShouldNotEqual, eTag1)
			})
		})

		Convey("getNewETagForRemoveDimension returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := newETagForRemoveDimension(currentFilter, testFilterID, testDimensionName)
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)

			Convey("Applying the same update to a different filter results in a different ETag", func() {
				filter2 := testFilter()
				filter2.FilterID = "otherFilter"
				eTag2, err := newETagForRemoveDimension(filter2, testFilterID, testDimensionName)
				So(err, ShouldBeNil)
				So(eTag2, ShouldNotEqual, eTag1)
			})

			Convey("Applying a different update to the same filter results in a different ETag", func() {
				eTag3, err := newETagForRemoveDimension(currentFilter, testFilterID, "otherDimension")
				So(err, ShouldBeNil)
				So(eTag3, ShouldNotEqual, eTag1)
			})
		})
	})
}

func TestGetNewETagForDimensionOptionsOperations(t *testing.T) {

	Convey("Given a filer with a dimension whose options we want to update", t, func() {

		currentFilter := testFilter()

		Convey("getNewETagForAddDimensionOptions returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := newETagForAddDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op4", "op5"})
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)

			Convey("Applying the same update to a different filter results in a different ETag", func() {
				filter2 := testFilter()
				filter2.FilterID = "otherFilter"
				eTag2, err := newETagForAddDimensionOptions(filter2, testFilterID, testDimensionName, []string{"op4", "op5"})
				So(err, ShouldBeNil)
				So(eTag2, ShouldNotEqual, eTag1)
			})

			Convey("Applying a different update to the same filter results in a different ETag", func() {
				eTag3, err := newETagForAddDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op4"})
				So(err, ShouldBeNil)
				So(eTag3, ShouldNotEqual, eTag1)
			})

			Convey("Removing the same dimensions from the same filter results in a different ETag", func() {
				eTag4, err := newETagForRemoveDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op4", "op5"})
				So(err, ShouldBeNil)
				So(eTag4, ShouldNotEqual, eTag1)
			})
		})

		Convey("getNewETagForRemoveDimensionOptions returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := newETagForRemoveDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op1", "op3"})
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)

			Convey("Applying the same update to a different filter results in a different ETag", func() {
				filter2 := testFilter()
				filter2.FilterID = "otherFilter"
				eTag2, err := newETagForRemoveDimensionOptions(filter2, testFilterID, testDimensionName, []string{"op1", "op3"})
				So(err, ShouldBeNil)
				So(eTag2, ShouldNotEqual, eTag1)
			})

			Convey("Applying a different update to the same filter results in a different ETag", func() {
				eTag3, err := newETagForRemoveDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op2"})
				So(err, ShouldBeNil)
				So(eTag3, ShouldNotEqual, eTag1)
			})

			Convey("Adding the same dimensions from the same filter results in a different ETag", func() {
				eTag4, err := newETagForAddDimensionOptions(currentFilter, testFilterID, testDimensionName, []string{"op1", "op3"})
				So(err, ShouldBeNil)
				So(eTag4, ShouldNotEqual, eTag1)
			})
		})
	})
}
