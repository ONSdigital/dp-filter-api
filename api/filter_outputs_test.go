package api

import (
	"testing"

	"encoding/json"
	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"strings"

	"context"
	"errors"
	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/go-ns/common"
	"net/http"
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
				buildDownloadsObject(option.inputPreviousFilterOutput, option.inputFilterOutput, downloadServiceURL)
				So(option.inputFilterOutput, ShouldResemble, option.expectedOutput)
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

	expectedAuditParams := common.Params{"filter_output_id": "12345678"}

	Convey("Successfully get a filter output from an unauthenticated request", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		// Check private link is hidden for unauthenticated user
		jsonResult := w.Body.Bytes()

		filterOutput := &models.Filter{}
		if err = json.Unmarshal(jsonResult, filterOutput); err != nil {
			t.Logf("failed to marshal filte output json response, error: [%v]", err.Error())
			t.Fail()
		}

		So(filterOutput.Downloads.CSV, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.csv", Private: "", Public: "", Size: "12mb"})
		So(filterOutput.Downloads.XLS, ShouldResemble, &models.DownloadItem{HRef: "ons-test-site.gov.uk/87654321.xls", Private: "", Public: "", Size: "24mb"})

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get a filter output from a request with an authorised download service token", t, func() {

		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		r.Header.Add(common.DownloadServiceHeaderKey, downloadServiceToken)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
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

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get an unpublished filter output", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterOutput(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "12345678"}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter output does not exist, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter output is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterOutput_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "12345678"}

	Convey("Given an existing published filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter outputs endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getFilterOutputAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter outputs endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getFilterOutputAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterOutputAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter outputs endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getFilterOutputAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulUpdateFilterOutput(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Successfully update filter output when public csv download link is missing", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully update filter output when public xls download link is missing", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "public":"s3-public-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestSuccessfulUpdateFilterOutputUnpublished(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Successfully update filter output with private csv download link when version is unpublished", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private": "s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully update filter output with private xls download link when version is unpublished", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToUpdateFilterOutput(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When an update to a filter output resource that does not exist, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a json message contains private csv link but current filter output has private csv download links already and version is published, than a forbidden status is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private":"s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.csv.private]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a json message contains private xls link but current filter output has private xls download links already and version is published, than a forbidden status is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.xls.private]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToUpdateFilterOutput_BadRequest(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Given an existing filter output with download links", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a PUT request is made to the filter output endpoint with invalid JSON", func() {
			reader := strings.NewReader("{")
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, badRequestResponse)
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with empty JSON", func() {
			reader := strings.NewReader("{}")
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, badRequestResponse)
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with fields that are not allowed to be updated", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "Forbidden from updating the following fields: [dataset.id dataset.edition dataset.version]\n")
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with the wrong authorisation header", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 401 unauthorised", func() {
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, filters.ErrUnauthorised.Error()+"\n")
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with contains a CSV download", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.csv]\n")
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with contains an XLS download", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"s3-xls-location","size":"12mb", "public":"s3-public-xls-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "Forbidden from updating the following fields: [downloads.xls]\n")
			})

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})
		})
	})
}

func TestUpdateFilterOutput_PrivateEndpointsNotEnabled(t *testing.T) {

	Convey("When private endpoints are not enabled, calling update on the filter output returns a 404 not found", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, false, downloadServiceURL, downloadServiceToken, getMockAuditor())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusMethodNotAllowed)
	})
}

func TestFailedToUpdateFilterOutput_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Given an existing filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{MissingPublicLinks: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a PUT request is made to the filter output endpoint and the attempt audit fails", func() {

			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], updateFilterOutputAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint and the outcome audit fails", func() {

			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == updateFilterOutputAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with invalid json and the outcome audit fails", func() {

			reader := strings.NewReader("{")
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == updateFilterOutputAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterOutputAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulGetPreview(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Successfully requesting a valid preview", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMock.GetPreviewCalls()[0].Limit, ShouldEqual, 20)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully requesting a valid preview for unpublished version filters", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMock.GetPreviewCalls()[0].Limit, ShouldEqual, 20)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully requesting a valid preview with a new limit", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=10", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		previewMockForLimit := &datastoretest.PreviewDatasetMock{
			GetPreviewFunc: func(filter *models.Filter, limit int) (*preview.FilterPreview, error) {
				return &preview.FilterPreview{}, nil
			},
		}
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMockForLimit, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(previewMockForLimit.GetPreviewCalls()[0].Limit, ShouldEqual, 10)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedGetPreview(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "21312"}

	Convey("Requesting a preview with invalid filter", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("Requesting a preview with no mongodb database connection", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("Requesting a preview with no neo4j database connection", t, func() {
		mockAuditor := getMockAuditor()
		previewMockInternalError := &datastoretest.PreviewDatasetMock{
			GetPreviewFunc: func(filter *models.Filter, limit int) (*preview.FilterPreview, error) {
				return nil, errors.New("internal error")
			},
		}
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMockInternalError, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, "internal server error\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("Requesting a preview with no dimensions", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "no dimensions are present in the filter\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("Requesting a preview with an invalid limit", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=a", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "requested limit is not a number\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("Requesting a preview with no authentication when the version is unpublished", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/21312/preview?limit=a", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "requested limit is not a number\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetPreview_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_output_id": "12345678"}

	Convey("Given an existing published filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter preview endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678/preview", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getFilterPreviewAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter preview endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678/preview", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getFilterPreviewAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterPreviewAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter preview endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678/preview", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getFilterPreviewAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterPreviewAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
