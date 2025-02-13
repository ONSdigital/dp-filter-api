package filteroutputqueue

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOuputQueue(t *testing.T) {
	Convey("When a filter output is created, a message is sent to kafka", t, func() {
		filterOutputQueue := make(chan []byte, 1)
		outputQueue := CreateOutputQueue(filterOutputQueue)
		dataset := &models.Dataset{
			ID:      "cpih01",
			Edition: "2017",
			Version: 1,
		}
		filter := models.Filter{InstanceID: "12345678", State: "submitted", FilterID: "1234", Dataset: dataset}
		err := outputQueue.Queue(&filter)
		So(err, ShouldBeNil)

		bytes := <-filterOutputQueue
		var filterMessage filterOutput
		schema.FilterSubmittedSchema.Unmarshal(bytes, &filterMessage)
		So(filterMessage.FilterOutputID, ShouldEqual, filter.FilterID)
	})
}
