package mongo

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateUpdateFilterOutput(t *testing.T) {
	Convey("When a filter output is updated with a new CSV file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				CSV: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.csv",
					Size: "321",
				},
			},
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.CSV.HRef, ShouldEndWith, filterOutput.Downloads.CSV.HRef)
			So(downloads.CSV.Size, ShouldEndWith, filterOutput.Downloads.CSV.Size)
		})
	})

	Convey("When a filter output is updated with a new XLSX file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				XLS: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.xlsx",
					Size: "3213",
				},
			},
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.XLS.HRef, ShouldEndWith, filterOutput.Downloads.XLS.HRef)
			So(downloads.XLS.Size, ShouldEndWith, filterOutput.Downloads.XLS.Size)
		})
	})

	Convey("When a filter output is updated with both a XLSX and CSV file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				XLS: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.xlsx",
					Size: "3213",
				},
				CSV: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.csv",
					Size: "321",
				},
			},
			State: models.CompletedState,
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			state := data["$set"].(bson.M)["state"].(string)
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.XLS.HRef, ShouldEndWith, filterOutput.Downloads.XLS.HRef)
			So(downloads.XLS.Size, ShouldEndWith, filterOutput.Downloads.XLS.Size)
			So(downloads.CSV.HRef, ShouldEndWith, filterOutput.Downloads.CSV.HRef)
			So(downloads.CSV.Size, ShouldEndWith, filterOutput.Downloads.CSV.Size)
			So(state, ShouldEqual, models.CompletedState)

		})
	})
}

func TestSelector(t *testing.T) {

	Convey("Given some testing values to provide as selector paramters", t, func() {
		var testFilterID string = "filterID"
		var testDimensionName string = "dimensionName"
		var testETag string = "testETag"
		var testMongoTimestamp bson.MongoTimestamp = 1234567890

		Convey("Then, providing an empty string dimension, zero timestamp and any eTag generates a selector that only queries by filter_id", func() {
			s := selector(testFilterID, "", 0, AnyETag)
			So(s, ShouldResemble, bson.M{"filter_id": testFilterID})
		})

		Convey("Then, providing values for dimension, timestamp, and eTag generates a selector that queries by filterID, dimensionName, timestamp and eTag", func() {
			s := selector(testFilterID, testDimensionName, testMongoTimestamp, testETag)
			So(s, ShouldResemble, bson.M{
				"filter_id":        testFilterID,
				"dimensions":       bson.M{"$elemMatch": bson.M{"name": testDimensionName}},
				"unique_timestamp": testMongoTimestamp,
				"e_tag":            testETag,
			})
		})
	})
}

func TestValidateFilter(t *testing.T) {
	Convey("validateFilter creates empty slices for dimensions and events if the values were nil", t, func() {
		filter := models.Filter{}
		validateFilter(&filter)
		So(filter.Dimensions, ShouldResemble, []models.Dimension{})
		So(filter.Events, ShouldResemble, []*models.Event{})
	})
}
