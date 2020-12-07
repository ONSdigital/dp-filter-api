package mocks

import (
	"errors"

	"github.com/globalsign/mgo/bson"

	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
)

// A list of errors that can be returned by mock package
var (
	errorInternalServer = errors.New("DataStore internal error")
)

// DataStoreConfig represents a list of error flags to set error in mocked datastore
type DataStoreConfig struct {
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

// DataStore holds the list of possible error flats along with the mocked datastore.
// This struct can be used directly as a mock, as it implements the required methods,
// or you can use the internal 'moq' Mock if you want ot validate calls, parameters etc.
type DataStore struct {
	Cfg  DataStoreConfig
	Mock *datastoretest.DataStoreMock
}

// NewDataStore creates a new datastore mock with an empty config
func NewDataStore() *DataStore {
	ds := &DataStore{
		Cfg: DataStoreConfig{},
	}
	mock := &datastoretest.DataStoreMock{
		AddFilterFunc:                    ds.AddFilter,
		AddFilterDimensionFunc:           ds.AddFilterDimension,
		AddFilterDimensionOptionFunc:     ds.AddFilterDimensionOption,
		AddFilterDimensionOptionsFunc:    ds.AddFilterDimensionOptions,
		CreateFilterOutputFunc:           ds.CreateFilterOutput,
		GetFilterFunc:                    ds.GetFilter,
		GetFilterDimensionFunc:           ds.GetFilterDimension,
		GetFilterOutputFunc:              ds.GetFilterOutput,
		RemoveFilterDimensionFunc:        ds.RemoveFilterDimension,
		RemoveFilterDimensionOptionFunc:  ds.RemoveFilterDimensionOption,
		RemoveFilterDimensionOptionsFunc: ds.RemoveFilterDimensionOptions,
		UpdateFilterFunc:                 ds.UpdateFilter,
		UpdateFilterOutputFunc:           ds.UpdateFilterOutput,
		AddEventToFilterOutputFunc:       ds.AddEventToFilterOutput,
	}
	ds.Mock = mock
	return ds
}

// NotFound sets NotFound flag to true
func (ds *DataStore) NotFound() *DataStore {
	ds.Cfg.NotFound = true
	return ds
}

// DimensionNotFound sets DimensionNotFound flag to true
func (ds *DataStore) DimensionNotFound() *DataStore {
	ds.Cfg.DimensionNotFound = true
	return ds
}

// OptionNotFound sets OptionNotFound flag to true
func (ds *DataStore) OptionNotFound() *DataStore {
	ds.Cfg.OptionNotFound = true
	return ds
}

// VersionNotFound sets VersionNotFound flag to true
func (ds *DataStore) VersionNotFound() *DataStore {
	ds.Cfg.VersionNotFound = true
	return ds
}

// ChangeInstanceRequest sets ChangeInstanceRequest flag to true
func (ds *DataStore) ChangeInstanceRequest() *DataStore {
	ds.Cfg.ChangeInstanceRequest = true
	return ds
}

// InvalidDimensionOption sets InvalidDimensionOption flag to true
func (ds *DataStore) InvalidDimensionOption() *DataStore {
	ds.Cfg.InvalidDimensionOption = true
	return ds
}

// Unpublished sets Unpublished flag to true
func (ds *DataStore) Unpublished() *DataStore {
	ds.Cfg.Unpublished = true
	return ds
}

// InternalError sets InternalError flag to true
func (ds *DataStore) InternalError() *DataStore {
	ds.Cfg.InternalError = true
	return ds
}

// MissingPublicLinks sets MissingPublicLinks flag to true
func (ds *DataStore) MissingPublicLinks() *DataStore {
	ds.Cfg.MissingPublicLinks = true
	return ds
}

// BadRequest sets BadRequest flag to true
func (ds *DataStore) BadRequest() *DataStore {
	ds.Cfg.BadRequest = true
	return ds
}

// ConflictRequest sets ConflictRequest flag to true
func (ds *DataStore) ConflictRequest() *DataStore {
	ds.Cfg.ConflictRequest = true
	return ds
}

// AgeDimension sets AgeDimension flag to true
func (ds *DataStore) AgeDimension() *DataStore {
	ds.Cfg.AgeDimension = true
	return ds
}

// AddFilter represents the mocked version of creating a filter blueprint to the datastore
func (ds *DataStore) AddFilter(filterJob *models.Filter) (*models.Filter, error) {
	if ds.Cfg.InternalError {
		return nil, errorInternalServer
	}
	return &models.Filter{InstanceID: "12345678"}, nil
}

// AddFilterDimension represents the mocked version of creating a filter dimension to the datastore
func (ds *DataStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// AddFilterDimensionOption represents the mocked version of creating a filter dimension option to the datastore
func (ds *DataStore) AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// AddFilterDimensionOptions represents the mocked version of adding a list of dimension options to the datastore
func (ds *DataStore) AddFilterDimensionOptions(filterID, name string, options []string, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrDimensionNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// CreateFilterOutput represents the mocked version of creating a filter output to the datastore
func (ds *DataStore) CreateFilterOutput(filterJob *models.Filter) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	return nil
}

// GetFilter represents the mocked version of getting a filter blueprint from the datastore
func (ds *DataStore) GetFilter(filterID string) (*models.Filter, error) {
	if ds.Cfg.NotFound {
		return nil, filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.InternalError {
		return nil, errorInternalServer
	}

	if ds.Cfg.BadRequest {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678"}, nil
	}

	if ds.Cfg.InvalidDimensionOption {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "age", Options: []string{"28"}}}}, nil
	}

	if ds.Cfg.Unpublished {
		if ds.Cfg.DimensionNotFound {
			return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}}}, nil
		}
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{Name: "age", Options: []string{"33"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/1_age", Name: "1_age", Options: []string{"2014", "2015"}}}}, nil
	}

	if ds.Cfg.DimensionNotFound {
		return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Dimensions: []models.Dimension{{URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}}}, nil
	}

	return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "age", Options: []string{"33"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/time", Name: "time", Options: []string{"2014", "2015"}}, {URL: "http://localhost:22100/filters/12345678/dimensions/1_age", Name: "1_age", Options: []string{"2014", "2015"}}}}, nil
}

// GetFilterDimension represents the mocked version of getting a filter dimension from the datastore
func (ds *DataStore) GetFilterDimension(filterID, name string) (*models.Dimension, error) {
	if ds.Cfg.DimensionNotFound {
		return nil, filters.ErrDimensionNotFound
	}

	if ds.Cfg.InternalError {
		return nil, errorInternalServer
	}

	return &models.Dimension{Name: "1_age"}, nil
}

// GetFilterOutput represents the mocked version of getting a filter output from the datastore
func (ds *DataStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	if ds.Cfg.NotFound {
		return nil, filters.ErrFilterOutputNotFound
	}

	if ds.Cfg.InternalError {
		return nil, errorInternalServer
	}

	if ds.Cfg.BadRequest {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "created"}, nil
	}

	if ds.Cfg.Unpublished {
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

	if ds.Cfg.MissingPublicLinks {
		return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "created", Dimensions: []models.Dimension{{Name: "time"}}, Downloads: downloads}, nil
	}

	downloads.CSV.Public = "csv-public-link"
	downloads.XLS.Public = "xls-public-link"

	return &models.Filter{InstanceID: "12345678", FilterID: filterID, Published: &models.Published, State: "completed", Dimensions: []models.Dimension{{Name: "time"}}, Downloads: downloads}, nil
}

// RemoveFilterDimension represents the mocked version of removing a filter dimension from the datastore
func (ds *DataStore) RemoveFilterDimension(string, string, bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// RemoveFilterDimensionOption represents the mocked version of removing a filter dimension option from the datastore
func (ds *DataStore) RemoveFilterDimensionOption(filterJobID, name, option string, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.DimensionNotFound {
		return filters.ErrDimensionNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// RemoveFilterDimensionOptions represents the mocked version of removing a set of filter dimension options from the datastore
func (ds *DataStore) RemoveFilterDimensionOptions(filterJobID, name string, options []string, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrDimensionNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// UpdateFilter represents the mocked version of updating a filter blueprint from the datastore
func (ds *DataStore) UpdateFilter(filterJob *models.Filter, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.VersionNotFound {
		return filters.ErrVersionNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterBlueprintConflict
	}

	return nil
}

// UpdateFilterOutput represents the mocked version of updating a filter output from the datastore
func (ds *DataStore) UpdateFilterOutput(filterJob *models.Filter, timestamp bson.MongoTimestamp) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterBlueprintNotFound
	}

	if ds.Cfg.ConflictRequest {
		return filters.ErrFilterOutputConflict
	}

	return nil
}

// AddEventToFilterOutput adds the given event to the filter output of the given ID
func (ds *DataStore) AddEventToFilterOutput(filterOutputID string, event *models.Event) error {
	if ds.Cfg.InternalError {
		return errorInternalServer
	}

	if ds.Cfg.NotFound {
		return filters.ErrFilterOutputNotFound
	}

	return nil
}
