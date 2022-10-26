package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-filter-api/api"
	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-filter-api/mock"
	dprequest "github.com/ONSdigital/dp-net/request"
)

const (
	filterID1 = "121"
	filterID2 = "122"
	filterID3 = "123"
)

func TestSuccessfulGetFilterOutput(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully get a filter output from an unauthenticated request", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
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
	})

	Convey("Successfully get a filter output from a request with an authorised download service token", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		r.Header.Add(dprequest.DownloadServiceHeaderKey, downloadServiceToken)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
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
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterOutput(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter output does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")
	})

	Convey("When filter output is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterOutputNotFound.Error()+"\n")
	})
}

func TestSuccessfulUpdateFilterOutput(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully update filter output when public csv download link is missing", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().MissingPublicLinks(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully update filter output when public xls download link is missing", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "public":"s3-public-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().MissingPublicLinks(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestSuccessfulUpdateFilterOutput_StatusComplete(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given a filter output without downloads", t, func() {
		mockDatastore := &apimock.DataStoreMock{
			AddEventToFilterOutputFunc: func(ctx context.Context, filterOutputID string, event *models.Event) error {
				return nil
			},
			GetFilterOutputFunc: func(ctx context.Context, filterOutputID string) (*models.Filter, error) {
				return createFilter(), nil
			},
			UpdateFilterOutputFunc: func(ctx context.Context, filterOutput *models.Filter, timestamp primitive.Timestamp) error {
				return nil
			},
		}

		filterApi := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When the PUT filter output endpoint is called with completed download data", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}, "xls":{"size":"12mb", "public":"s3-public-xls-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to update the event", func() {
				So(len(mockDatastore.UpdateFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.UpdateFilterOutputCalls()[0].Filter
				So(filterOutput.State, ShouldEqual, models.CompletedState)
			})

			Convey("Then the data store is called to add a completed event", func() {
				So(len(mockDatastore.AddEventToFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.AddEventToFilterOutputCalls()[0]
				So(filterOutput.Event.Type, ShouldEqual, models.EventFilterOutputCompleted)
			})

			Convey("Then the response code should be 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When the PUT filter output endpoint is called with completed csv but skipped xls download data", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}, "xls":{"skipped":true}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to update the event", func() {
				So(len(mockDatastore.UpdateFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.UpdateFilterOutputCalls()[0].Filter
				So(filterOutput.State, ShouldEqual, models.CompletedState)
			})

			Convey("Then the data store is called to add a completed event", func() {
				So(len(mockDatastore.AddEventToFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.AddEventToFilterOutputCalls()[0]
				So(filterOutput.Event.Type, ShouldEqual, models.EventFilterOutputCompleted)
			})

			Convey("Then the response code should be 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulUpdateFilterOutputUnpublished(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully update filter output with private csv download link when version is unpublished", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private": "s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully update filter output with private xls download link when version is unpublished", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterOutput(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When an update to a filter output resource that does not exist, a not found is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message contains private csv link but current filter output has private csv download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "private":"s3-private-csv-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().MissingPublicLinks(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "forbidden from updating the following fields: [downloads.csv.private]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message contains private xls link but current filter output has private xls download links already and version is published, than a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"xls":{"size":"12mb", "private":"s3-private-xls-location"}}}`)
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().MissingPublicLinks(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		response := w.Body.String()
		So(response, ShouldResemble, "forbidden from updating the following fields: [downloads.xls.private]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterOutput_BadRequest(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an existing filter output with download links", t, func() {
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When a PUT request is made to the filter output endpoint with invalid JSON", func() {
			reader := strings.NewReader("{")
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, badRequestResponse)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with empty JSON", func() {
			reader := strings.NewReader("{}")
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, badRequestResponse)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with fields that are not allowed to be updated", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "forbidden from updating the following fields: [dataset.id dataset.edition dataset.version]\n")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with the wrong authorisation header", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
			So(err, ShouldBeNil)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 401 unauthorised", func() {
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, filters.ErrUnauthorised.Error()+"\n")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with contains a CSV download", func() {
			reader := strings.NewReader(`{"downloads":{"csv":{"size":"12mb", "public":"s3-public-csv-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "forbidden from updating the following fields: [downloads.csv]\n")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filter output endpoint with contains an XLS download", func() {
			reader := strings.NewReader(`{"downloads":{"xls":{"href":"s3-xls-location","size":"12mb", "public":"s3-public-xls-location"}}}`)
			r := createAuthenticatedRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)

			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 403 forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Then the response contains the expected content", func() {
				response := w.Body.String()
				So(response, ShouldResemble, "forbidden from updating the following fields: [downloads.xls]\n")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulAddEventToFilterOutput(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an existing filter output", t, func() {

		mockDatastore := &apimock.DataStoreMock{
			AddEventToFilterOutputFunc: func(ctx context.Context, filterOutputID string, event *models.Event) error {
				return nil
			},
			GetFilterOutputFunc: func(ctx context.Context, filterOutputID string) (*models.Filter, error) {
				return createFilter(), nil
			},
		}

		filterApi := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When a POST request is made to the filter output event endpoint", func() {

			reader := strings.NewReader(`{"type":"` + models.EventFilterOutputCompleted + `","time":"2018-06-10T05:59:05.893629647+01:00"}`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filter-outputs/21312/events", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to add the event", func() {
				So(len(mockDatastore.AddEventToFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.AddEventToFilterOutputCalls()[0]
				So(filterOutput.Event.Type, ShouldEqual, models.EventFilterOutputCompleted)
			})

			Convey("Then the response is 201 OK", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedAddEventToFilterOutput_InvalidJson(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an existing filter output", t, func() {

		mockDatastore := &apimock.DataStoreMock{}

		filterApi := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When a POST request is made to the filter output event endpoint with invalid json", func() {

			reader := strings.NewReader(`{`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filter-outputs/21312/events", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedAddEventToFilterOutput_InvalidEvent(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an existing filter output", t, func() {

		mockDatastore := &apimock.DataStoreMock{}

		filterApi := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When a POST request is made to the filter output event endpoint with an empty event type", func() {

			reader := strings.NewReader(`{"type":""}`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filter-outputs/21312/events", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedAddEventToFilterOutput_DatastoreError(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an existing filter output", t, func() {

		mockDatastore := &apimock.DataStoreMock{
			AddEventToFilterOutputFunc: func(ctx context.Context, filterOutputID string, event *models.Event) error {
				return errors.New("database is broken")
			},
			GetFilterOutputFunc: func(ctx context.Context, filterOutputID string) (*models.Filter, error) {
				return createFilter(), nil
			},
		}

		filterApi := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock)

		Convey("When a POST request is made to the filter output event endpoint, and the data store returns an error", func() {

			reader := strings.NewReader(`{"type":"` + models.EventFilterOutputCompleted + `","time":"2018-06-10T05:59:05.893629647+01:00"}`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filter-outputs/21312/events", reader)

			w := httptest.NewRecorder()
			filterApi.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to add the event", func() {
				So(len(mockDatastore.AddEventToFilterOutputCalls()), ShouldEqual, 1)
				filterOutput := mockDatastore.AddEventToFilterOutputCalls()[0]
				So(filterOutput.Event.Type, ShouldEqual, models.EventFilterOutputCompleted)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func createFilter() *models.Filter {
	downloads := &models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    "ons-test-site.gov.uk/87654321.csv",
			Private: "csv-private-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			HRef:    "ons-test-site.gov.uk/87654321.xls",
			Private: "xls-private-link",
			Size:    "24mb",
		},
	}
	return &models.Filter{InstanceID: "12345678", FilterID: "543", Published: &models.Published, State: "created", Dimensions: []models.Dimension{{Name: "time"}}, Downloads: downloads}
}
