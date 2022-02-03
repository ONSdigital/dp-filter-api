package middleware

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
)

type datasetAPIClient interface{
	Get(ctx context.Context, userToken, svcToken, collectionID, datasetID string) (dataset.DatasetDetails, error)
}

type filterFlexAPIClient interface{
	ForwardRequest(*http.Request) (*http.Response, error)
}
