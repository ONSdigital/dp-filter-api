package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
)

//go:generate moq -out datastoretest/datastore.go -pkg datastoretest . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(host string, filter *models.Filter) (*models.Filter, error)
	AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension) error
	AddFilterDimensionOption(filterID, name, option string) error
	CreateFilterOutput(filter *models.Filter) error
	GetFilter(filterID string) (*models.Filter, error)
	GetFilterDimension(filterID string, name string) error
	GetFilterOutput(filterOutputID string) (*models.Filter, error)
	RemoveFilterDimension(filterID string, name string) error
	RemoveFilterDimensionOption(filterID string, name string, option string) error
	UpdateFilter(filter *models.Filter) error
	UpdateFilterOutput(filterOutput *models.Filter) error
}
