package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

var internalServerError = errors.New("DataStore internal error")

type DataStore struct {
	NotFound      bool
	BadRequest    bool
	InternalError bool
}

func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (models.Filter, error) {
	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DataSet: "Census", Version: "1", Edition: "1"}, nil
}
