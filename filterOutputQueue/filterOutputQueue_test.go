package filterOutputQueue

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOuputQueue(t *testing.T) {
	Convey("When a filter output is created, a message is sent to kafka", t, func() {
		filterOutputQueue := make(chan kafka.BytesMessage, 1)
		outputQueue := CreateOutputQueue(filterOutputQueue)
		dataset := &models.Dataset{
			ID:      "cpih01",
			Edition: "2017",
			Version: 1,
		}
		filter := models.Filter{InstanceID: "12345678", State: "submitted", FilterID: "1234", Dataset: dataset}
		err := outputQueue.Queue(context.Background(), &filter)
		So(err, ShouldBeNil)

		message := <-filterOutputQueue
		var filterMessage filterOutput
		schema.FilterSubmittedSchema.Unmarshal(message.Value, &filterMessage)
		So(filterMessage.FilterOutputID, ShouldEqual, filter.FilterID)
	})
}
