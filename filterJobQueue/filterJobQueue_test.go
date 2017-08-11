package filterJobQueue

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterJobQueue(t *testing.T) {
	Convey("When a filter job is updated with a status `submitted`, a message is sent to kafka", t, func() {
		filterJobQueue := make(chan []byte, 1)
		jobQueue := CreateJobQueue(filterJobQueue)
		filter := models.Filter{DatasetFilterID: "12345678", State: "submitted", FilterID: "1234"}
		err := jobQueue.Queue(&filter)
		So(err, ShouldBeNil)
		bytes := <-filterJobQueue
		var filterMessage filterJob
		schema.FilterJobSubmittedSchema.Unmarshal(bytes, &filterMessage)
		So(filterMessage.FilterJobID, ShouldEqual, filter.FilterID)
	})
}
