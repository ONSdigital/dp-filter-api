package filterOutputQueue

import (
	"context"
	"strconv"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
)

// Output is an object containng the filter output queue channel
type Output struct {
	FilterOutputQueue chan kafka.BytesMessage
}

type filterOutput struct {
	FilterOutputID string `avro:"filter_output_id"`
	DatasetID      string `avro:"dataset_id"`
	Edition        string `avro:"edition"`
	Version        string `avro:"version"`
}

// CreateOutputQueue returns an object containing a channel for queueing filter outputs
func CreateOutputQueue(filterOutputQueue chan kafka.BytesMessage) Output {
	return Output{FilterOutputQueue: filterOutputQueue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (filter *Output) Queue(ctx context.Context, outputFilter *models.Filter) error {
	message := filterOutput{
		FilterOutputID: outputFilter.FilterID,
		DatasetID:      outputFilter.Dataset.ID,
		Edition:        outputFilter.Dataset.Edition,
		Version:        strconv.Itoa(outputFilter.Dataset.Version),
	}
	bytes, err := schema.FilterSubmittedSchema.Marshal(message)
	if err != nil {
		return err
	}

	filter.FilterOutputQueue <-  kafka.BytesMessage{Value: bytes, Context: ctx} 

	return nil
}
