package mocks

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
)

type DatasetAPI struct {
	DimensionsNotFound       bool
	DimensionOptionsNotFound bool
	InstanceNotFound         bool
	InternalServerError      bool
}

var (
	instanceNotFoundError         = errors.New("Instance not found")
	dimensionsNotFoundError       = errors.New("Dimensions not found")
	dimensionOptionsNotFoundError = errors.New("Dimension options not found")
)

func (ds *DatasetAPI) GetInstance(ctx context.Context, id string) (*models.Instance, error) {
	if ds.InternalServerError {
		return nil, internalServerError
	}

	if ds.InstanceNotFound {
		return nil, instanceNotFoundError
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
	}, nil
}

func (ds *DatasetAPI) GetVersionDimensions(ctx context.Context, datasetID, edition, version string) (*models.DatasetDimensionResults, error) {
	if ds.InternalServerError {
		return nil, internalServerError
	}

	if ds.DimensionsNotFound {
		return nil, dimensionsNotFoundError
	}

	dimension := models.DatasetDimension{
		Name: "age",
	}

	return &models.DatasetDimensionResults{
		Items: []models.DatasetDimension{dimension},
	}, nil
}

func (ds *DatasetAPI) GetVersionDimensionOptions(ctx context.Context, datasetID, edition, version, dimension string) (*models.DatasetDimensionOptionResults, error) {
	if ds.InternalServerError {
		return nil, internalServerError
	}

	if ds.DimensionOptionsNotFound {
		return nil, dimensionOptionsNotFoundError
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
