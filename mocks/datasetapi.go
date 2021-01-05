package mocks

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	apimocks "github.com/ONSdigital/dp-filter-api/api/mocks"
	"github.com/ONSdigital/dp-filter-api/filters"
)

// DatasetAPIConfig represents a list of error flags to set error in mocked dataset API
type DatasetAPIConfig struct {
	VersionNotFound     bool
	InternalServerError bool
	Unpublished         bool
}

// DatasetAPI holds the list of possible error flats along with the mocked dataset api calls.
// This struct can be used directly as a mock, as it implements the required methods,
// or you can use the internal 'moq' Mock if you want ot validate calls, parameters etc.
type DatasetAPI struct {
	Cfg  DatasetAPIConfig
	Mock *apimocks.DatasetAPIMock
}

// NewDatasetAPI creates a new dataset API mock with an empty config
func NewDatasetAPI() *DatasetAPI {
	ds := &DatasetAPI{
		Cfg: DatasetAPIConfig{},
	}
	ds.Mock = &apimocks.DatasetAPIMock{
		GetVersionFunc:             ds.GetVersion,
		GetVersionDimensionsFunc:   ds.GetVersionDimensions,
		GetOptionsBatchProcessFunc: ds.GetOptionsBatchProcess,
	}
	return ds
}

// VersionNotFound sets VersionNotFound flag to true
func (ds *DatasetAPI) VersionNotFound() *DatasetAPI {
	ds.Cfg.VersionNotFound = true
	return ds
}

// InternalServiceError sets InternalServiceError flag to true
func (ds *DatasetAPI) InternalServiceError() *DatasetAPI {
	ds.Cfg.InternalServerError = true
	return ds
}

// Unpublished sets Unpublished flag to true
func (ds *DatasetAPI) Unpublished() *DatasetAPI {
	ds.Cfg.Unpublished = true
	return ds
}

// GetVersion represents the mocked version of getting an version document from dataset API
func (ds *DatasetAPI) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error) {
	if ds.Cfg.InternalServerError {
		return m, errorInternalServer
	}

	if ds.Cfg.VersionNotFound {
		return m, filters.ErrVersionNotFound
	}

	if ds.Cfg.Unpublished {
		return dataset.Version{
			Links: dataset.Links{
				Dataset: dataset.Link{
					ID: "123",
				},
				Edition: dataset.Link{
					ID: "2017",
				},
				Version: dataset.Link{
					ID: "1",
				},
			},
		}, nil
	}

	return dataset.Version{
		Links: dataset.Links{
			Dataset: dataset.Link{
				ID: "123",
			},
			Edition: dataset.Link{
				ID: "2017",
			},
			Version: dataset.Link{
				ID: "1",
			},
		},
		State: "published",
	}, nil
}

// GetVersionDimensions represents the mocked version of getting a list of dimensions from the dataset API
func (ds *DatasetAPI) GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error) {
	if ds.Cfg.InternalServerError {
		return m, errorInternalServer
	}

	dimension := dataset.VersionDimension{
		Name: "age",
	}

	return dataset.VersionDimensions{
		Items: []dataset.VersionDimension{dimension},
	}, nil
}

// GetOptionsBatchProcess represents the mocked version of getting a list of dimension options from the dataset API
func (ds *DatasetAPI) GetOptionsBatchProcess(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, optionIDs *[]string, processBatch dataset.OptionsBatchProcessor, batchSize int, maxWorkers int) (err error) {

	if ds.Cfg.InternalServerError {
		return errorInternalServer
	}

	dimensionOptionOne := dataset.Option{
		Label:  "age",
		Option: "27",
	}

	dimensionOptionTwo := dataset.Option{
		Label:  "age",
		Option: "33",
	}

	offset := 0
	for offset < 2 {

		// get items for the offset
		items := slice([]dataset.Option{dimensionOptionOne, dimensionOptionTwo}, offset, batchSize)
		opts := dataset.Options{
			Items:      items,
			TotalCount: 2,
			Offset:     offset,
			Limit:      batchSize,
			Count:      len(items),
		}

		// call the provided processor
		abort, err := processBatch(opts)
		if abort || err != nil {
			return err
		}

		// increase offset for next iteration
		offset += batchSize
	}

	return nil
}

func slice(full []dataset.Option, offset, limit int) (sliced []dataset.Option) {
	end := offset + limit
	if limit == 0 || end > len(full) {
		end = len(full)
	}

	if offset > len(full) {
		return []dataset.Option{}
	}

	return full[offset:end]
}
