package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

var internalServerError = errors.New("DataStore internal error")
var badRequestError = errors.New("Bad request")
var forbiddenError = errors.New("Forbidden")
var notFoundError = errors.New("Not found")

type DataStore struct {
	NotFound      bool
	BadRequest    bool
	Forbidden     bool
	Unauthorised  bool
	InternalError bool
}

func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (models.Filter, error) {
	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DataSetFilterID: "12345678"}, nil
}

func (ds *DataStore) UpdateFilter(host string, filterJob *models.Filter) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	if ds.NotFound {
		return notFoundError
	}
	return nil
}
