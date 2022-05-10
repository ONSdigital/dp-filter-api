package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/mongo.go -pkg mock . MongoDB

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// MongoDB defines the required methods from MongoDB package
type MongoDB interface {
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
	GetFilterOutput(ctx context.Context, filterID string) (*models.Filter, error)
	GetCantabularFilterOutput(ctx context.Context, filterOutputID string) (*models.Filter, error)
	UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error
	AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Close(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}
