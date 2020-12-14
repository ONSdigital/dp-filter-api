package api

import (
	"io"
	"net/http"

	"context"

	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	dprequest "github.com/ONSdigital/dp-net/request"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	host                   = "http://localhost:80"
	maxRequestOptions      = 1000
	enablePrivateEndpoints = true
	downloadServiceURL     = "http://localhost:23600"
	downloadServiceToken   = "123123"
	serviceAuthToken       = "321321"
)

var (
	filterNotFoundResponse        = filters.ErrFilterBlueprintNotFound.Error() + "\n"
	dimensionNotFoundResponse     = filters.ErrDimensionNotFound.Error() + "\n"
	versionNotFoundResponse       = filters.ErrVersionNotFound.Error() + "\n"
	optionNotFoundResponse        = filters.ErrDimensionOptionNotFound.Error() + "\n"
	invalidQueryParameterResponse = filters.ErrInvalidQueryParameter.Error() + "\n"
	badRequestResponse            = badRequest + "\n"
	internalErrResponse           = internalError + "\n"
)

// cfg obtains a new config for testing. Each test will have its own config instance by using this func.
func cfg() *config.Config {
	return &config.Config{
		Host:                     host,
		MaxRequestOptions:        maxRequestOptions,
		DownloadServiceURL:       downloadServiceURL,
		DownloadServiceSecretKey: downloadServiceToken,
		ServiceAuthToken:         serviceAuthToken,
		EnablePrivateEndpoints:   enablePrivateEndpoints,
		DatasetLimit:             1000,
		MongoConfig: config.MongoConfig{
			Limit:  0,
			Offset: 0,
		},
	}
}

var previewMock = &datastoretest.PreviewDatasetMock{
	GetPreviewFunc: func(ctx context.Context, filter *models.Filter, limit int) (*models.FilterPreview, error) {
		return &models.FilterPreview{}, nil
	},
}

func createAuthenticatedRequest(method, url string, body io.Reader) *http.Request {

	r, err := http.NewRequest(method, url, body)
	ctx := r.Context()
	ctx = dprequest.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)

	So(err, ShouldBeNil)
	return r
}
