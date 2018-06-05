package mongo

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gedge/mgo/bson"
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
	Convey("When a filter output is updated with event and state", t, func() {
		filterOutput := models.Filter{
			State: models.CompletedState,
			Events: []*models.Event{{
				Type: "wut",
				Time: time.Now().UTC(),
			}},
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			state := data["$set"].(bson.M)["state"].(string)
			events := data["$set"].(bson.M)["events"].([]*models.Event)
			So(len(events), ShouldEqual, 1)
			So(state, ShouldEqual, models.CompletedState)
		})
	})
}
