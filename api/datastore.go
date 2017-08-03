package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
)

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(host string, filter *models.Filter) (models.Filter, error)
	UpdateFilter(host string, filter *models.Filter) error
}
