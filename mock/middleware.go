package mock

import (
	"bytes"
	"context"
	"io"
	"net/http"

	apimock "github.com/ONSdigital/dp-filter-api/api/mock"

	"github.com/ONSdigital/dp-filter-api/models"
)

func GenerateMocksForMiddleware(flexCode int, datasetVersion int, filterType string) (*apimock.FilterFlexAPIMock, *apimock.DataStoreMock) {
	// set the Filter flex Call
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{
		ForwardRequestFunc: func(r *http.Request) (*http.Response, error) { //
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
				StatusCode: flexCode,
			}, nil
		},
	}

	datastoreMock := NewDataStore().Mock

	datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
		return &models.Filter{
			Dataset: &models.Dataset{
				Version: datasetVersion,
			},
			Type: filterType,
		}, nil
	}

	datastoreMock.GetFilterOutputFunc = func(ctx context.Context, filterID string) (*models.Filter, error) {
		return &models.Filter{
			Dataset: &models.Dataset{
				Version: datasetVersion,
			},
			Type: filterType,
		}, nil
	}

	return filterFlexAPIMock, datastoreMock
}
