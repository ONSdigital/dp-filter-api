package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/globalsign/mgo/bson"
)

//go:generate moq -out datastoretest/datastore.go -pkg datastoretest . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(filter *models.Filter) (*models.Filter, error)
	AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp) error
	AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp) error
	AddFilterDimensionOptions(filterID, name string, options []string, timestamp bson.MongoTimestamp) (int, error)
	CreateFilterOutput(filter *models.Filter) error
	GetFilter(filterID string) (*models.Filter, error)
	GetFilterDimension(filterID string, name string) (*models.Dimension, error)
	GetFilterOutput(filterOutputID string) (*models.Filter, error)
	RemoveFilterDimension(filterID string, name string, timestamp bson.MongoTimestamp) error
	RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp) error
	RemoveFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp) (int, error)
	UpdateFilter(filter *models.Filter, timestamp bson.MongoTimestamp) error
	UpdateFilterOutput(filterOutput *models.Filter, timestamp bson.MongoTimestamp) error
	AddEventToFilterOutput(filterOutputID string, event *models.Event) error
}
