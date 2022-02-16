package middleware

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"

	"github.com/pkg/errors"
)

type datasetAPIClient interface {
	Get(ctx context.Context, userToken, svcToken, collectionID, datasetID string) (dataset.DatasetDetails, error)
}

type filterFlexAPIClient interface {
	ForwardRequest(*http.Request) (*http.Response, error)
}

type dataLogger interface {
	LogData() map[string]interface{}
}

type coder interface {
	Code() int
}

type messager interface {
	Message() string
}

type stacktracer interface {
	StackTrace() errors.StackTrace
}

type responder interface {
	JSON(context.Context, http.ResponseWriter, int, interface{})
	Error(context.Context, http.ResponseWriter, int, error)
}
