package api

import (
	"io"
	"net/http"

	"context"
	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	host                   = "http://localhost:80"
	enablePrivateEndpoints = true
	downloadServiceURL     = "http://localhost:23600"
	downloadServiceToken   = "123wut"
)

var (
	filterNotFoundResponse    = filters.ErrFilterBlueprintNotFound.Error() + "\n"
	dimensionNotFoundResponse = filters.ErrDimensionNotFound.Error() + "\n"
	versionNotFoundResponse   = filters.ErrVersionNotFound.Error() + "\n"
	optionNotFoundResponse    = filters.ErrDimensionOptionNotFound.Error() + "\n"
	badRequestResponse        = badRequest + "\n"
	internalErrResponse       = internalError + "\n"
)

var previewMock = &datastoretest.PreviewDatasetMock{
	GetPreviewFunc: func(filter *models.Filter, limit int) (*preview.FilterPreview, error) {
		return &preview.FilterPreview{}, nil
	},
}

func createAuthenticatedRequest(method, url string, body io.Reader) *http.Request {

	r, err := http.NewRequest(method, url, body)
	ctx := r.Context()
	ctx = common.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)

	So(err, ShouldBeNil)
	return r
}

func getMockAuditor() *audit.AuditorServiceMock {
	return &audit.AuditorServiceMock{
		RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
			return nil
		},
	}
}
