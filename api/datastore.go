package api

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-filter-api/models"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

//go:generate moq -out mock/datastore.go -pkg mock . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(ctx context.Context, filter *models.Filter) (*models.Filter, error)
	GetFilter(ctx context.Context, filterID, eTagSelector string) (*models.Filter, error)
	UpdateFilter(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	GetFilterDimension(ctx context.Context, filterID string, name, eTagSelector string) (dimension *models.Dimension, err error)
	AddFilterDimension(ctx context.Context, filterID, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimension(ctx context.Context, filterID, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOption(ctx context.Context, filterID, name, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOptions(ctx context.Context, filterID, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	CreateFilterOutput(ctx context.Context, filter *models.Filter) error
	GetFilterOutput(ctx context.Context, filterOutputID string) (*models.Filter, error)
	UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error
	AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error
	RunTransaction(ctx context.Context, fn mongodriver.SessionFunc) error
}
