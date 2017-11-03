package filterOutputQueue

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
		filter := models.Filter{InstanceID: "12345678", State: "submitted", FilterID: "1234"}
		err := outputQueue.Queue(&filter)
		So(err, ShouldBeNil)

		bytes := <-filterOutputQueue
		var filterMessage filterOutput
		schema.FilterSubmittedSchema.Unmarshal(bytes, &filterMessage)
		So(filterMessage.FilterOutputID, ShouldEqual, filter.FilterID)
	})
}
