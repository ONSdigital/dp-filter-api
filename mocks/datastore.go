package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

// A list of errors that can be returned by mock package
var (
	errorInternalServer       = errors.New("DataStore internal error")
	errorUnauthorised         = errors.New("Unauthorised")
	errorBadRequest           = errors.New("Bad request")
	errorForbidden            = errors.New("Forbidden")
	errorNotFound             = errors.New("Not found")
	errorDimensionionNotFound = errors.New("Dimension not found")
	errorOptionNotFound       = errors.New("Option not found")
	errorFilterOutputNotFound = errors.New("Filter output not found")
)

// DataStore represents a list of error flags to set error in mocked datastore
type DataStore struct {
	NotFound               bool
	DimensionNotFound      bool
	OptionNotFound         bool
	InstanceNotFound       bool
	BadRequest             bool
	Forbidden              bool
	Unauthorised           bool
	InternalError          bool
	ChangeInstanceRequest  bool
	InvalidDimensionOption bool
	Unpublished            bool
}

// AddFilter represents the mocked version of creating a filter blueprint to the datastore
func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (*models.Filter, error) {
	if ds.InternalError {
		return nil, errorInternalServer
	}
	return &models.Filter{InstanceID: "12345678"}, nil
}

// AddFilterDimension represents the mocked version of creating a filter dimension to the datastore
func (ds *DataStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return errorNotFound
	}

	if ds.Forbidden {
		return errorForbidden
	}

	return nil
}

// AddFilterDimensionOption represents the mocked version of creating a filter dimension option to the datastore
func (ds *DataStore) AddFilterDimensionOption(filterID, name, option string) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return errorNotFound
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.Forbidden {
		return errorForbidden
	}

	return nil
}

// CreateFilterOutput represents the mocked version of creating a filter output to the datastore
func (ds *DataStore) CreateFilterOutput(filterJob *models.Filter) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.Unauthorised {
		return errorUnauthorised
	}
	return nil
}

// GetFilter represents the mocked version of getting a filter blueprint from the datastore
func (ds *DataStore) GetFilter(filterID string) (*models.Filter, error) {
	if ds.NotFound {
		return nil, errorNotFound
	}

	if ds.InternalError {
		return nil, errorInternalServer
	}

	if ds.BadRequest {
		return nil, errorBadRequest
	}

	if ds.ChangeInstanceRequest {
		return &models.Filter{FilterID: filterID, InstanceID: "12345678", Published: true, Dimensions: []models.Dimension{{Name: "age", Options: []string{"33"}}}}, nil
	}

	if ds.InvalidDimensionOption {
		return &models.Filter{FilterID: filterID, InstanceID: "12345678", Published: true, Dimensions: []models.Dimension{{Name: "age", Options: []string{"28"}}}}, nil
	}

	if ds.Unpublished {
		return &models.Filter{FilterID: filterID, InstanceID: "12345678", Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}}}, nil

	}

	return &models.Filter{FilterID: filterID, InstanceID: "12345678", Published: true, Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}}}, nil
}

// GetFilterDimension represents the mocked version of getting a filter dimension from the datastore
func (ds *DataStore) GetFilterDimension(filterID, name string) error {
	if ds.DimensionNotFound {
		return errorDimensionionNotFound
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.InternalError {
		return errorInternalServer
	}

	return nil
}

// GetFilterOutput represents the mocked version of getting a filter output from the datastore
func (ds *DataStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	if ds.NotFound {
		return nil, errorFilterOutputNotFound
	}

	if ds.InternalError {
		return nil, errorInternalServer
	}

	if ds.BadRequest {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: true, State: "created"}, nil
	}

	if ds.Unpublished {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, State: "created", Dimensions: []models.Dimension{{Name: "time"}}, Links: models.LinkMap{FilterBlueprint: models.LinkObject{ID: filterID}}}, nil

	}

	return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: true, State: "created", Dimensions: []models.Dimension{{Name: "time"}}}, nil
}

// RemoveFilterDimension represents the mocked version of removing a filter dimension from the datastore
func (ds *DataStore) RemoveFilterDimension(string, string) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.Forbidden {
		return errorForbidden
	}

	if ds.NotFound {
		return errorNotFound
	}

	return nil
}

// RemoveFilterDimensionOption represents the mocked version of removing a filter dimension option from the datastore
func (ds *DataStore) RemoveFilterDimensionOption(filterJobID, name, option string) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.Forbidden {
		return errorForbidden
	}

	if ds.DimensionNotFound {
		return errorDimensionionNotFound
	}

	return nil
}

// UpdateFilter represents the mocked version of updating a filter blueprint from the datastore
func (ds *DataStore) UpdateFilter(filterJob *models.Filter) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.NotFound {
		return errorFilterOutputNotFound
	}

	if ds.InstanceNotFound {
		return errorInstanceNotFound
	}
	return nil
}

// UpdateFilterOutput represents the mocked version of updating a filter output from the datastore
func (ds *DataStore) UpdateFilterOutput(filterJob *models.Filter) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.BadRequest {
		return errorBadRequest
	}

	if ds.NotFound {
		return errorNotFound
	}

	return nil
}
