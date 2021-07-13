package mock

import (
	"fmt"

	"github.com/ONSdigital/dp-filter-api/models"
)

// FilterJob contains a flag indicating whether the job has failed or not
type FilterJob struct {
	ReturnError bool
}

// MessageData contains the unique identifier for the filter job
type MessageData struct {
	FilterJobID string
}

// Queue checks whether the filter job has errored
func (fj *FilterJob) Queue(filter *models.Filter) error {
	if fj.ReturnError {
		return fmt.Errorf("No message produced for filter job")
	}
	return nil
}
