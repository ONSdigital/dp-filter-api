package filterJobQueue

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/schema"
)

type job struct {
	filterJobQueue chan []byte
}

type filterJob struct {
	FilterJobID string `avro:"filter_job_id"`
}

// CreateJobQueue returns an object containing a channel for queueing filter jobs
func CreateJobQueue(filterJobQueue chan []byte) job {
	return job{filterJobQueue: filterJobQueue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (filter *job) Queue(jobFilter *models.Filter) error {
	message := filterJob{FilterJobID: jobFilter.FilterID}
	bytes, err := schema.FilterJobSubmittedSchema.Marshal(message)
	if err != nil {
		return err
	}

	filter.filterJobQueue <- bytes

	return nil
}
