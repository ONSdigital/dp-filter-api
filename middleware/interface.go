package middleware

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
)

type datasetAPIClient interface {
	Get(ctx context.Context, userToken, svcToken, collectionID, datasetID string) (dataset.DatasetDetails, error)
}

type filterFlexAPIClient interface {
	ForwardRequest(*http.Request) (*http.Response, error)
}

type datastore interface {
	GetFilter(ctx context.Context, filterID, eTagSelector string) (*models.Filter, error)
	GetFilterOutput(ctx context.Context, filterID string) (*models.Filter, error)
}

type responder interface {
	JSON(context.Context, http.ResponseWriter, int, interface{})
	Error(context.Context, http.ResponseWriter, int, error)
}
