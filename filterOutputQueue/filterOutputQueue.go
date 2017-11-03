package filterOutputQueue

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
)

type output struct {
	filterOutputQueue chan []byte
}

type filterOutput struct {
	FilterOutputID string `avro:"filter_output_id"`
}

// CreateOutputQueue returns an object containing a channel for queueing filter outputs
func CreateOutputQueue(filterOutputQueue chan []byte) output {
	return output{filterOutputQueue: filterOutputQueue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (filter *output) Queue(outputFilter *models.Filter) error {
	message := filterOutput{FilterOutputID: outputFilter.FilterID}
	bytes, err := schema.FilterSubmittedSchema.Marshal(message)
	if err != nil {
		return err
	}

	filter.filterOutputQueue <- bytes

	return nil
}
