package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

var (
	internalServerError  = errors.New("DataStore internal error")
	badRequestError      = errors.New("Bad request")
	forbiddenError       = errors.New("Forbidden")
	notFoundError        = errors.New("Not found")
	dimensionionNotFound = errors.New("Dimension not found")
	optionNotFound       = errors.New("Option not found")
)

type DataStore struct {
	NotFound          bool
	DimensionNotFound bool
	OptionNotFound    bool
	BadRequest        bool
	Forbidden         bool
	Unauthorised      bool
	InternalError     bool
}

func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (models.Filter, error) {
	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DataSetFilterID: "12345678"}, nil
}

func (ds *DataStore) AddFilterDimension(dimension *models.AddDimension) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.NotFound {
		return notFoundError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	return nil
}

func (ds *DataStore) AddFilterDimensionOption(dimension *models.AddDimensionOption) error {
	if ds.InternalError {
		return internalServerError
	}

	if ds.NotFound {
		return notFoundError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.Forbidden {
		return forbiddenError
	}

	return nil
}

func (ds *DataStore) GetFilter(filterID string) (models.Filter, error) {
	if ds.NotFound {
		return models.Filter{}, notFoundError
	}

	if ds.InternalError {
		return models.Filter{}, internalServerError
	}
	return models.Filter{DataSetFilterID: "12345678"}, nil
}

func (ds *DataStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	dimensions := []models.Dimension{}

	if ds.NotFound {
		return dimensions, notFoundError
	}

	if ds.InternalError {
		return dimensions, internalServerError
	}

	dimensions = append(dimensions, models.Dimension{Name: "1_age", DimensionURL: "/filters/123/dimensions/1_age"})

	return dimensions, nil
}

func (ds *DataStore) GetFilterDimension(filterID string, name string) error {
	if ds.NotFound {
		return notFoundError
	}

	if ds.BadRequest {
		return badRequestError
	}

	if ds.InternalError {
		return internalServerError
	}

	return nil
}

func (ds *DataStore) GetFilterDimensionOptions(filterID string, name string) (models.GetDimensionOptions, error) {
	var (
		options    models.GetDimensionOptions
		optionURLs []string
	)

	if ds.BadRequest {
		return options, badRequestError
	}

	if ds.DimensionNotFound {
		return options, dimensionionNotFound
	}

	if ds.InternalError {
		return options, internalServerError
	}

	optionURLs = append(optionURLs, "/filters/123/dimensions/1_age/options/26")

	options.DimensionOptionURLs = optionURLs

	return options, nil
}

func (ds *DataStore) GetFilterDimensionOption(filterID string, name string, option string) error {
	if ds.BadRequest {
		return badRequestError
	}

	if ds.OptionNotFound {
		return optionNotFound
	}

	if ds.InternalError {
		return internalServerError
	}

	return nil
}

func (ds *DataStore) RemoveFilterDimension(string, string) error {
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
