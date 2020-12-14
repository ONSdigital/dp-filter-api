package mocks

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-filter-api/filters"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	VersionNotFound     bool
	InternalServerError bool
	Unpublished         bool
}

// GetVersion represents the mocked version of getting an version document from dataset API
func (ds *DatasetAPI) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error) {
	if ds.InternalServerError {
		return m, errorInternalServer
	}

	if ds.VersionNotFound {
		return m, filters.ErrVersionNotFound
	}

	if ds.Unpublished {
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
	if ds.InternalServerError {
		return m, errorInternalServer
	}

	dimension := dataset.VersionDimension{
		Name: "age",
	}

	return dataset.VersionDimensions{
		Items: []dataset.VersionDimension{dimension},
	}, nil
}

// GetOptions represents the mocked version of getting a list of dimension options from the dataset API
func (ds *DatasetAPI) GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, offset, limit int) (m dataset.Options, err error) {
	if ds.InternalServerError {
		return m, errorInternalServer
	}

	dimensionOptionOne := dataset.Option{
		Label:  "age",
		Option: "27",
	}

	dimensionOptionTwo := dataset.Option{
		Label:  "age",
		Option: "33",
	}

	items := slice([]dataset.Option{dimensionOptionOne, dimensionOptionTwo}, offset, limit)

	return dataset.Options{
		Items:      items,
		TotalCount: 2,
		Offset:     offset,
		Limit:      limit,
		Count:      len(items),
	}, nil
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
