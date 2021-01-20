package mongo

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func testStore() *FilterStore {
	return &FilterStore{
		host:              "testHost",
		db:                "filters",
		filtersCollection: "filters",
		outputsCollection: "filterOutputs",
	}
}

func testFilter() *models.Filter {
	notPublished := false
	f := &models.Filter{
		InstanceID: "testInstanceID",
		FilterID:   "testFilterID",
		State:      models.CreatedState,
		Published:  &notPublished,
	}
	eTag0, err := f.Hash()
	So(err, ShouldBeNil)
	f.ETag = eTag0
	return f
}

func TestGetNewETag(t *testing.T) {

	Convey("Given a FilterStore instance for testing", t, func() {

		store := testStore()
		currentFilter := testFilter()

		published := true
		update := &models.Filter{
			FilterID:  "testFilterID",
			State:     models.CreatedState,
			Published: &published,
		}

		Convey("getNewETagForUpdate returns an eTag that is different from the original filter ETag", func() {
			eTag1, err := store.getNewETagForUpdate(nil, 0, "", currentFilter, update)
			So(err, ShouldBeNil)
			So(eTag1, ShouldNotEqual, currentFilter.ETag)
		})
	})

}
