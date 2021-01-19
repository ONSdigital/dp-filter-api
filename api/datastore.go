package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/globalsign/mgo/bson"
)

//go:generate moq -out mocks/datastore.go -pkg mocks . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(filter *models.Filter) (*models.Filter, error)
	GetFilter(filterID, eTagSelector string) (*models.Filter, error)
	UpdateFilter(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	GetFilterDimension(filterID string, name, eTagSelector string) (dimension *models.Dimension, err error)
	AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimension(filterID, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOptions(filterID, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	CreateFilterOutput(filter *models.Filter) error
	GetFilterOutput(filterOutputID string) (*models.Filter, error)
	UpdateFilterOutput(filter *models.Filter, timestamp bson.MongoTimestamp) error
	AddEventToFilterOutput(filterOutputID string, event *models.Event) error
}
