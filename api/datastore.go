package api

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
)

//go:generate moq -out mocks/datastore.go -pkg mocks . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(ctx context.Context, filter *models.Filter) (*models.Filter, error)
	GetFilter(ctx context.Context, filterID, eTagSelector string) (*models.Filter, error)
	UpdateFilter(ctx context.Context, updatedFilter *models.Filter, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	GetFilterDimension(ctx context.Context, filterID string, name, eTagSelector string) (dimension *models.Dimension, err error)
	AddFilterDimension(ctx context.Context, filterID, name string, options []string, dimensions []models.Dimension, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimension(ctx context.Context, filterID, name string, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOption(ctx context.Context, filterID, name, option string, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOptions(ctx context.Context, filterID, name string, options []string, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp int64, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	CreateFilterOutput(ctx context.Context, filter *models.Filter) error
	GetFilterOutput(ctx context.Context, filterOutputID string) (*models.Filter, error)
	UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp int64) error
	AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error
}
