package api

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
)

// DatasetAPIer - An interface used to access the DatasetAPI
type DatasetAPIer interface {
	GetInstance(ctx context.Context, instanceID string) (*models.Instance, error)
	GetVersionDimensions(ctx context.Context, datasetID, edition, version string) (*models.DatasetDimensionResults, error)
	GetVersionDimensionOptions(ctx context.Context, datasetID, edition, version, dimension string) (*models.DatasetDimensionOptionResults, error)
}
