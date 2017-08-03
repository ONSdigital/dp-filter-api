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

func CreateJobQueue(filterJobQueue chan []byte) job {
	return job{filterJobQueue: filterJobQueue}
}

func (filter *job) Queue(jobFilter *models.Filter) error {
	message := filterJob{FilterJobID: jobFilter.FilterID}
	bytes, err := schema.FilterJobSubmittedSchema.Marshal(message)
	if err != nil {
		return err
	}

	filter.filterJobQueue <- bytes

	return nil
}
