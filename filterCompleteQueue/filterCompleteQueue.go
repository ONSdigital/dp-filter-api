package filterCompleteQueue

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
)

// Completed is an object containing the complete filter job queue channel
type Completed struct {
	FilterCompleteQueue chan []byte
}

type filterComplete struct {
	FilterCompleteID string `avro:"filter_complete_id"`
	Email            string `avro:"email"`
}

// CreateCompleteQueue returns an object containing a channel for queueing completed filter jobs
func CreateCompleteQueue(filterCompleteQueue chan []byte) Completed {
	return Completed{FilterCompleteQueue: filterCompleteQueue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (filter *Completed) Queue(completeFilter *models.Filter) error {
	message := filterComplete{
		FilterCompleteID: completeFilter.FilterID,
		Email:            completeFilter.Email,
	}
	bytes, err := schema.FilterCompletedEvent.Marshal(message)
	if err != nil {
		return err
	}

	filter.FilterCompleteQueue <- bytes

	return nil
}
