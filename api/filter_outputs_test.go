package api

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/ONSdigital/dp-filter-api/filters"
	"strings"
	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/preview"

	"net/http"
	"github.com/ONSdigital/dp-filter-api/mocks"
	"errors"
	"github.com/ONSdigital/go-ns/common"
)

const (
	filterID1 = "121"
	filterID2 = "122"
	filterID3 = "123"
)

var testOptions = []struct {
	inputPreviousFilterOutput *models.Filter
	inputFilterOutput         *models.Filter
	expectedOutput            *models.Filter
	title                     string
}{
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &fullDownloads},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "1",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &fullDownloads},
		inputFilterOutput:         &models.Filter{Downloads: nil},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "2",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &csvDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &csvDownloadsOnlyOutputs},
		title:                     "3",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &xlsDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &xlsDownloadsOnlyOutputs},
		title:                     "4",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "5",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "6",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[1].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[0].csv}},
		title:                     "7",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[2].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[1].csv}},
		title:                     "8",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv}},
		title:                     "9",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[1].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[3].xls}},
		title:                     "10",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[2].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[4].xls}},
		title:                     "11",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[5].xls}},
		title:                     "12",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv, XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv, XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv, XLS: &expectedDownloadItems[5].xls}},
		title:                     "13",
	},
}

func TestBuildDownloadsObject(t *testing.T) {

	Convey("Successfully build download object", t, func() {

		for _, option := range testOptions {
			Convey(option.title, func() {
				filterOutput := buildDownloadsObject(option.inputPreviousFilterOutput, option.inputFilterOutput, downloadServiceURL)
				So(filterOutput, ShouldResemble, option.expectedOutput)
			})
		}
	})
}

// Test data
var (
	fullDownloads = models.Downloads{
		CSV: &models.DownloadItem{
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	}

	csvDownloadsOnly = models.Downloads{
		CSV: &models.DownloadItem{
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: nil,
	}

	xlsDownloadsOnly = models.Downloads{
		XLS: &models.DownloadItem{
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
		CSV: nil,
	}

	fullDownloadsOutputs = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	}

	csvDownloadsOnlyOutputs = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: nil,
	}

	xlsDownloadsOnlyOutputs = models.Downloads{
		XLS: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
		CSV: nil,
	}
)

var xlsScenario = []struct {
	xls models.DownloadItem
}{
	{
		xls: models.DownloadItem{
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		xls: models.DownloadItem{
			Private: "xls-private-downloads-link-2",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
			Public: "xls-public-downloads-link-3",
			Size:   "44mb",
		},
	},
	{
		xls: models.DownloadItem{
			Public: "xls-public-downloads-link-4",
		},
	},
}

var csvScenario = []struct {
	csv models.DownloadItem
}{
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
	},
	{
		csv: models.DownloadItem{
			Private: "csv-private-downloads-link-2",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
			Public: "csv-public-downloads-link-3",
			Size:   "34mb",
		},
	},
	{
		csv: models.DownloadItem{
			Public: "csv-public-downloads-link-4",
		},
	},
}

var expectedDownloadItems = []struct {
	csv models.DownloadItem
	xls models.DownloadItem
}{
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID2 + ".csv",
			Private: "csv-private-downloads-link-2",
			Public:  "csv-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID3 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-3",
			Size:    "34mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-4",
			Size:    "12mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID2 + ".xlsx",
			Private: "xls-private-downloads-link-2",
			Public:  "xls-public-downloads-link",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID3 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-3",
			Size:    "44mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-4",
			Size:    "24mb",
		},
	},
}

func TestSuccessfulGetFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a filter output from an unauthenticated request", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		// Check private link is hidden for unauthenticated user
		jsonResult := w.Body.Bytes()

		filterOutput := &models.Filter{}
		if err = json.Unmarshal(jsonResult, filterOutput); err != nil {
			t.Logf("failed to marshal filte output json response, error: [%v]", err.Error())
			t.Fail()
		}

		So(filterOutput.Downloads.CSV, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.csv", Private: "", Public: "csv-public-link", Size: "12mb"})
		So(filterOutput.Downloads.XLS, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.xls", Private: "", Public: "xls-public-link", Size: "24mb"})
	})

	Convey("Successfully get a filter output from a request with an authorised download service token", t, func() {

		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		r.Header.Add(common.DownloadServiceHeaderKey, downloadServiceToken)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		// Check private link is NOT hidden from authenticated user
		jsonResult := w.Body.Bytes()

		filterOutput := &models.Filter{}
		if err := json.Unmarshal(jsonResult, filterOutput); err != nil {
			t.Logf("failed to marshal filte output json response, error: [%v]", err.Error())
			t.Fail()
		}

		So(filterOutput.Downloads.CSV, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.csv", Private: "csv-private-link", Public: "csv-public-link", Size: "12mb"})
		So(filterOutput.Downloads.XLS, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.xls", Private: "xls-private-link", Public: "xls-public-link", Size: "24mb"})
	})

	Convey("Successfully get an unpublished filter output", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/1234568", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter output does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")
	})

	Convey("When filter output is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")
	})
}

func TestSuccessfulUpdateFilterOutput(t *testing.T) {
	t.Parallel()

	Convey("Successfully update filter output when public csv download link is missing", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully update filter output when public xls download link is missing", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "public":"s3-public-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestSuccessfulUpdateFilterOutputUnpublished(t *testing.T) {
	t.Parallel()

	Convey("Successfully update filter output with private csv download link when version is unpublished", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private": "s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully update filter output with private xls download link when version is unpublished", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToUpdateFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)
	})

	Convey("When an update to a filter output resource that does not exist, a not found is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)
	})

	Convey("When a json message contains fields that are not allowed to be updated, a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [dataset.id dataset.edition dataset.version]\n")
	})

	Convey("When a json message is sent to change a filter output with the wrong authorisation header, an unauthorised status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)

		response := w.Body.String()
		So(response, ShouldResemble, errUnauthorised.Error()+"\n")
	})

	Convey("When a json message contains downloads object but current filter ouput has public csv download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.csv]\n")
	})

	Convey("When a json message contains downloads object but current filter ouput has public xls download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"href":"s3-xls-location","size":"12mb", "public":"s3-public-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.xls]\n")
	})

	Convey("When a json message contains private csv link but current filter ouput has private csv download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private":"s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.csv.private]\n")
	})

	Convey("When a json message contains private xls link but current filter ouput has private xls download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.xls.private]\n")
	})
}

func TestUpdateFilterOutput_PrivateEndpointsNotEnabled(t *testing.T) {

	Convey("When private endpoints are not enabled, calling update on the filter output returns a 404 not found", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, false, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusMethodNotAllowed)
	})
}

func TestSuccessfulGetPreview(t *testing.T) {
	t.Parallel()
	Convey("Successfully requesting a valid preview", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMock.GetPreviewCalls()[0].Limit, ShouldEqual, 20)
	})

	Convey("Successfully requesting a valid preview for unpublished version filters", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMock.GetPreviewCalls()[0].Limit, ShouldEqual, 20)
	})

	Convey("Successfully requesting a valid preview with a new limit", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=10", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		previewMockForLimit := &datastoretest.PreviewDatasetMock{
			GetPreviewFunc: func(filter *models.Filter, limit int) (*preview.FilterPreview, error) {
				return &preview.FilterPreview{}, nil
			},
		}
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMockForLimit, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMockForLimit.GetPreviewCalls()[0].Limit, ShouldEqual, 10)
	})
}

func TestFailedGetPreview(t *testing.T) {
	t.Parallel()
	Convey("Requesting a preview with invalid filter", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")
	})

	Convey("Requesting a preview with no mongodb database connection", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("Requesting a preview with no neo4j database connection", t, func() {
		previewMockInternalError := &datastoretest.PreviewDatasetMock{
			GetPreviewFunc: func(filter *models.Filter, limit int) (*preview.FilterPreview, error) {
				return nil, errors.New("internal error")
			},
		}
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMockInternalError, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, "internal server error\n")
	})

	Convey("Requesting a preview with no dimensions", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "no dimensions are present in the filter\n")
	})

	Convey("Requesting a preview with an invalid limit", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=a", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "requested limit is not a number\n")
	})

	Convey("Requesting a preview with no authentication when the version is unpublished", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=a", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "requested limit is not a number\n")
	})
}
