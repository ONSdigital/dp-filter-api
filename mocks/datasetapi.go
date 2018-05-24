package mocks

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	VersionNotFound     bool
	InternalServerError bool
	Unpublished         bool
}

// GetVersion represents the mocked version of getting an version document from dataset API
func (ds *DatasetAPI) GetVersion(ctx context.Context, dataset models.Dataset) (*models.Version, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
	}

	if ds.VersionNotFound {
		return nil, filters.ErrVersionNotFound
	}

	if ds.Unpublished {
		return &models.Version{
			Links: models.VersionLinks{
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

	return &models.Version{
		Links: models.VersionLinks{
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
func (ds *DatasetAPI) GetVersionDimensions(ctx context.Context, dataset models.Dataset) (*models.DatasetDimensionResults, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
	}

	dimension := models.DatasetDimension{
		Name: "age",
	}

	return &models.DatasetDimensionResults{
		Items: []models.DatasetDimension{dimension},
	}, nil
}

// GetVersionDimensionOptions represents the mocked version of getting a list of dimension options from the dataset API
func (ds *DatasetAPI) GetVersionDimensionOptions(ctx context.Context, dataset models.Dataset, dimension string) (*models.DatasetDimensionOptionResults, error) {
	if ds.InternalServerError {
		return nil, errorInternalServer
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
