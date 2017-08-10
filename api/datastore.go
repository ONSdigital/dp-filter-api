package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
)

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(host string, filter *models.Filter) (models.Filter, error)
	AddFilterDimension(*models.AddDimension) error
	AddFilterDimensionOption(*models.AddDimensionOption) error
	GetFilter(filterID string) (models.Filter, error)
	GetFilterDimensions(filterID string) ([]models.Dimension, error)
	GetFilterDimension(filterID string, name string) error
	GetFilterDimensionOptions(filterID string, name string) (models.GetDimensionOptions, error)
	GetFilterDimensionOption(filterID string, name string, option string) error
	RemoveFilterDimension(filterJobID string, name string) error
	UpdateFilter(host string, filter *models.Filter) error
}
