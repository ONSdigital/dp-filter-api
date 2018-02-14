package mocks

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	DimensionsNotFound       bool
	DimensionOptionsNotFound bool
	InstanceNotFound         bool
	InternalServerError      bool
}

// A list of errors that can be returned by the mock package
var (
	errorInstanceNotFound         = errors.New("Instance not found")
	errorDimensionsNotFound       = errors.New("Dimensions not found")
	errorDimensionOptionsNotFound = errors.New("Dimension options not found")
)

// GetInstance represents the mocked version of getting an instance document from dataset API
func (ds *DatasetAPI) GetInstance(ctx context.Context, id string) (*models.Instance, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
	}

	if ds.InstanceNotFound {
		return nil, errorInstanceNotFound
	}

	return &models.Instance{
		Links: models.InstanceLinks{
			Dataset: models.LinkObject{
				ID: "123",
			},
			Edition: models.LinkObject{
				ID: "2017",
			},
			Version: models.LinkObject{
				ID: "1",
			},
		},
		State: "published",
	}, nil
}

// GetVersionDimensions represents the mocked version of getting a list of dimensions from the dataset API
func (ds *DatasetAPI) GetVersionDimensions(ctx context.Context, datasetID, edition, version string) (*models.DatasetDimensionResults, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
	}

	if ds.DimensionsNotFound {
		return nil, errorDimensionsNotFound
	}

	dimension := models.DatasetDimension{
		Name: "age",
	}

	return &models.DatasetDimensionResults{
		Items: []models.DatasetDimension{dimension},
	}, nil
}

// GetVersionDimensionOptions represents the mocked version of getting a list of dimension options from the dataset API
func (ds *DatasetAPI) GetVersionDimensionOptions(ctx context.Context, datasetID, edition, version, dimension string) (*models.DatasetDimensionOptionResults, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
	}

	if ds.DimensionOptionsNotFound {
		return nil, errorDimensionOptionsNotFound
	}

	dimensionOptionOne := models.PublicDimensionOption{
		Name:   "age",
		Label:  "age",
		Option: "27",
	}

	dimensionOptionTwo := models.PublicDimensionOption{
		Name:   "age",
		Label:  "age",
		Option: "33",
	}

	return &models.DatasetDimensionOptionResults{
		Items: []models.PublicDimensionOption{dimensionOptionOne, dimensionOptionTwo},
	}, nil
}
