package mocks

import (
	"errors"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gedge/mgo/bson"
)

// A list of errors that can be returned by mock package
var (
	errorInternalServer = errors.New("DataStore internal error")
)

// DataStore represents a list of error flags to set error in mocked datastore
type DataStore struct {
	NotFound               bool
	DimensionNotFound      bool
	OptionNotFound         bool
	VersionNotFound        bool
	InternalError          bool
	ChangeInstanceRequest  bool
	InvalidDimensionOption bool
	Unpublished            bool
	MissingPublicLinks     bool
	BadRequest             bool
	ConflictRequest        bool
	AgeDimension           bool
}

// AddFilter represents the mocked version of creating a filter blueprint to the datastore
func (ds *DataStore) AddFilter(host string, filterJob *models.Filter) (*models.Filter, error) {
	if ds.InternalError {
		return nil, errorInternalServer
	}
	return &models.Filter{InstanceID: "12345678"}, nil
}

// AddFilterDimension represents the mocked version of creating a filter dimension to the datastore
func (ds *DataStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// AddFilterDimensionOption represents the mocked version of creating a filter dimension option to the datastore
func (ds *DataStore) AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// CreateFilterOutput represents the mocked version of creating a filter output to the datastore
func (ds *DataStore) CreateFilterOutput(filterJob *models.Filter) error {
	if ds.InternalError {
		return errorInternalServer
	}

	return nil
}

// GetFilter represents the mocked version of getting a filter blueprint from the datastore
func (ds *DataStore) GetFilter(filterID string) (*models.Filter, error) {
	if ds.NotFound {
		return nil, filters.ErrFilterBlueprintNotFound
	}

	if ds.InternalError {
		return nil, errorInternalServer
	}

	if ds.BadRequest {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678"}, nil
	}

	if ds.InvalidDimensionOption {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "age", Options: []string{"28"}}}}, nil
	}

	if ds.Unpublished {
		if ds.DimensionNotFound {
			return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}}}, nil
		}
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/1_age", Name: "1_age", Options: []string{"2014", "2015"}}}}, nil
	}

	if ds.DimensionNotFound {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}}}, nil
	}

	return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/1_age", Name: "1_age", Options: []string{"2014", "2015"}}}}, nil
}

// GetFilterDimension represents the mocked version of getting a filter dimension from the datastore
func (ds *DataStore) GetFilterDimension(filterID, name string) (*models.Dimension, error) {
	if ds.DimensionNotFound {
		return nil, filters.ErrDimensionNotFound
	}

	if ds.InternalError {
		return nil, errorInternalServer
	}

	return &models.Dimension{Name: "time", URL: "http://localhost:22100/filters/12345678/dimensions/1_age", Options: []string{"2014", "2015"}}, nil
}

// GetFilterOutput represents the mocked version of getting a filter output from the datastore
func (ds *DataStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	if ds.NotFound {
		return nil, filters.ErrFilterOutputNotFound
	}

	if ds.InternalError {
		return nil, errorInternalServer
	}

	if ds.BadRequest {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "created"}, nil
	}

	if ds.Unpublished {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, State: "created", Dimensions: []models.Dimension{{Name: "time"}}, Links: models.LinkMap{FilterBlueprint: models.LinkObject{ID: filterID}}}, nil
	}

	downloads := &models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    "ons-test-site.gov.uk/87654321.csv",
			Private: "csv-private-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			HRef:    "ons-test-site.gov.uk/87654321.xls",
			Private: "xls-private-link",
			Size:    "24mb",
		},
	}

	if ds.MissingPublicLinks {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "created", Dimensions: []models.Dimension{{Name: "time"}}, Downloads: downloads}, nil
	}

	downloads.CSV.Public = "csv-public-link"
	downloads.XLS.Public = "xls-public-link"

	return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "completed", Dimensions: []models.Dimension{{Name: "time"}}, Downloads: downloads}, nil
}

// RemoveFilterDimension represents the mocked version of removing a filter dimension from the datastore
func (ds *DataStore) RemoveFilterDimension(string, string, bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// RemoveFilterDimensionOption represents the mocked version of removing a filter dimension option from the datastore
func (ds *DataStore) RemoveFilterDimensionOption(filterJobID, name, option string, timestamp bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.DimensionNotFound {
		return filters.ErrDimensionNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// UpdateFilter represents the mocked version of updating a filter blueprint from the datastore
func (ds *DataStore) UpdateFilter(filterJob *models.Filter, timestamp bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.VersionNotFound {
		return filters.ErrVersionNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// UpdateFilterOutput represents the mocked version of updating a filter output from the datastore
func (ds *DataStore) UpdateFilterOutput(filterJob *models.Filter, timestamp bson.MongoTimestamp) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.ConflictRequest {
		return filters.ErrFilterOutputConflict
	}

	return nil
}

// AddEventToFilterOutput adds the given event to the filter output of the given ID
func (ds *DataStore) AddEventToFilterOutput(filterOutputID string, event *models.Event) error {
	if ds.InternalError {
		return errorInternalServer
	}

	if ds.NotFound {
		return filters.ErrFilterOutputNotFound
	}

	return nil
}
