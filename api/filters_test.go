package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-filter-api/api"
	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	testETag      = fmt.Sprintf("%s0", mock.TestETag)
	testETag1     = fmt.Sprintf("%s1", mock.TestETag)
	testETag2     = fmt.Sprintf("%s2", mock.TestETag)
	hostURL       = &url.URL{Scheme: "http", Host: "localhost:80"}
	datasetAPIURL = &url.URL{Scheme: "http", Host: "localhost:22000"}
)

func TestSuccessfulAddFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			AddFilterFunc: func(ctx context.Context, filter *models.Filter) (*models.Filter, error) {
				filter.ETag = testETag
				return filter, nil
			},
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
		}

		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a POST request is made to the filters endpoint", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with valid dimensions", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with the submitted query string parameter", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters?submitted=true", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, models.EventFilterOutputCreated)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulPostFilterBlueprint_Flexible(t *testing.T) {
	Convey("Given an flexible filter", t, func() {
		w := httptest.NewRecorder()
		filterFlexAPIMock := &apimock.FilterFlexAPIMock{
			ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
					StatusCode: http.StatusOK,
				}, nil
			},
		}
		datastoreMock := mock.NewDataStore().Mock
		datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
			return &models.Filter{
				Type: "flexible",
			}, nil
		}
		filterAPI := api.Setup(cfg(), mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a post request is made to the submit endpoint", func() {
			reader := strings.NewReader(`{}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters/123456/submit", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the request is not be forwarded to the filter flex api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})

			Convey("Then an error is returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestUnSuccessfulPostFilterBlueprint_NonFlexible(t *testing.T) {
	Convey("Given an non flexible filter", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true
		w := httptest.NewRecorder()
		filterFlexAPIMock := &apimock.FilterFlexAPIMock{
			ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
					StatusCode: http.StatusOK,
				}, nil
			},
		}
		datastoreMock := mock.NewDataStore().Mock
		datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
			return &models.Filter{
				Type: "flexible",
			}, nil
		}
		filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a post request is made to the submit endpoint", func() {
			reader := strings.NewReader(`{}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters/123456/submit", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 1)

				Convey("The request body is forwarded on", func() {
					call := filterFlexAPIMock.ForwardRequestCalls()[0]
					reqBody, err := io.ReadAll(call.Request.Body)
					So(err, ShouldBeNil)
					So(string(reqBody), ShouldEqual, `{}`)
				})
			})
		})
	})
}

func TestSuccessfulAddFilterBlueprint_UnpublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an unpublished dataset", t, func() {
		ds := mock.NewDataStore().Unpublished().Mock
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a POST request is made to the filters endpoint", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
				So(ds.AddFilterCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestFailedToAddFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When duplicate dimensions are sent", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"},"dimensions":[{"name":"time","options":["Jun-15","Jun-12"]},{"name":"time","options":["Jun-14"]}]}`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 400 bad request, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			response := w.Body.String()
			So(response, ShouldContainSubstring, "Bad request - duplicate dimension found: time")
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no data store is available", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When dataset API is unavailable", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().InternalServiceError(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version does not exist", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 404 Not Found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version is unpublished and the request is not authenticated", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 404 not found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToAddFilterBlueprint_BadJSON(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a POST request is made to the filters endpoint which has an invalid JSON message", func() {
			reader := strings.NewReader("{")
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which has an empty JSON message", func() {
			reader := strings.NewReader("{}")
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which is missing mandatory fields", func() {
			reader := strings.NewReader(`{"dataset":"Census"}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension that does not exist", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "weight", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimensions chosen: [weight]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension option that does not exist", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "age", "options": ["29","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimension options chosen: [29]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_PublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, true)

		Convey("When a GET request is made to the filters endpoint with no authentication", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 200 ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_PublishedDataset_URLRewriting(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}
	t.Parallel()

	Convey("Given a published dataset with URL rewriting enabled", t, func() {
		w := httptest.NewRecorder()
		mockDatastore := &apimock.DataStoreMock{
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
				if eTagSelector != mongo.AnyETag && eTagSelector != testETag {
					return nil, filters.ErrFilterBlueprintConflict
				}
				return &models.Filter{
					Dataset:    &models.Dataset{ID: "123", Edition: "2019", Version: 1},
					InstanceID: "12345678",
					Published:  &models.Published,
					Dimensions: []models.Dimension{
						{Name: "time", Options: []string{"2014", "2015"}},
						{Name: "1_age"},
					},
					ETag: testETag,
					Links: models.LinkMap{
						Self:       &models.LinkObject{HRef: fmt.Sprintf("%s/filters/%s", hostURL, filterID)},
						Dimensions: &models.LinkObject{HRef: fmt.Sprintf("%s/filters/%s/dimensions", hostURL, filterID)},
						Version:    &models.LinkObject{HRef: fmt.Sprintf("%s/datasets/123/editions/2019/versions/1", hostURL)},
					},
				}, nil
			},
		}
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, true)

		Convey("When a GET request is made to the filters endpoint without X-Forwarded-Host", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)
			bodyBytes := w.Body.Bytes()

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the rewritten links should be correct", func() {
				var responseBody map[string]interface{}
				err := json.Unmarshal(bodyBytes, &responseBody)
				So(err, ShouldBeNil)

				links, ok := responseBody["links"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				selfLink, ok := links["self"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(selfLink["href"], ShouldEqual, "http://localhost:80/filters/12345678")

				dimensionsLink, ok := links["dimensions"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(dimensionsLink["href"], ShouldEqual, "http://localhost:80/filters/12345678/dimensions")

				versionLink, ok := links["version"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(versionLink["href"], ShouldEqual, "http://localhost:22000/datasets/123/editions/2019/versions/1")
			})
		})

		Convey("When a GET request is made to the filters endpoint with X-Forwarded-Host", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			r.Header.Set("X-Forwarded-Host", "api.test.com")
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)
			bodyBytes := w.Body.Bytes()

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the rewritten links should be correct", func() {
				var responseBody map[string]interface{}
				err := json.Unmarshal(bodyBytes, &responseBody)
				So(err, ShouldBeNil)
				links, ok := responseBody["links"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				selfLink, ok := links["self"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(selfLink["href"], ShouldEqual, "https://api.test.com/filters/12345678")

				dimensionsLink, ok := links["dimensions"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(dimensionsLink["href"], ShouldEqual, "https://api.test.com/filters/12345678/dimensions")

				versionLink, ok := links["version"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(versionLink["href"], ShouldEqual, "https://api.test.com/datasets/123/editions/2019/versions/1")
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_UnpublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Successfully get an unpublished filter blueprint with authentication", t, func() {
		r := createAuthenticatedRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the expected ETag is returned in the ETag header", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
		})
	})
}

func TestFailedToGetFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished, and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})
}

func TestSuccessfulUpdateFilterBlueprint_PublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Given a published dataset", t, func() {
		testETag := "testETag"
		testETagUpdated := "testETagUpdated"
		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
				if eTagSelector != mongo.AnyETag && eTagSelector != testETag {
					return nil, filters.ErrFilterBlueprintConflict
				}
				return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}, {Name: "1_age"}}, ETag: testETag}, nil
			},
			UpdateFilterFunc: func(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
				if eTagSelector != testETag {
					return "", filters.ErrFilterBlueprintConflict
				}
				return testETagUpdated, nil
			},
		}

		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

		Convey("When a PUT request is made to the filters endpoint and a valid ETag", func() {
			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint with events and dataset version update", func() {
			updateBlueprintData := `{"dataset":{"version":1}, "events":[{"type":"wut","time":"2018-06-05T11:34:35.291735535+01:00"}]}`

			reader := strings.NewReader(updateBlueprintData)
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint to submit a filter blueprint", func() {
			reader := strings.NewReader("{}")
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312?submitted=true", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, models.EventFilterOutputCreated)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulUpdateFilterBlueprint_UnpublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Successfully send a request to submit an unpublished filter blueprint", t, func() {
		reader := strings.NewReader("{}")
		r := createAuthenticatedRequest("PUT", cfg().Host+"/filters/21312?submitted=true", reader)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When an empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to update filter blueprint that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no authentication is provided to update an unpublished filter, a not found is returned", t, func() {
		reader := strings.NewReader(`{"dimensions":[]}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the version does not exist, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound(), filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, versionNotFoundResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the datset version of a filter blueprint and the current dimensions do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [time 1_age]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the current dimension options do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [28]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a request is made without providing an If-Match header, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "required If-Match header not provided\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestRemoveDuplicatesAndEmptyOptions(t *testing.T) {
	Convey("Given a string array with duplicate options", t, func() {
		duplicates := []string{"1", "2", "2", "2", "abcde", "abd", "abcde"}

		Convey("When I call remove duplicate and empty function", func() {
			withoutDuplicates := api.RemoveDuplicateAndEmptyOptions(duplicates)

			Convey("Then the duplicates are removed", func() {
				expected := []string{"1", "2", "abcde", "abd"}
				So(withoutDuplicates, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a string array with empty options", t, func() {
		duplicates := []string{"", "1", "", "2", "", "3"}

		Convey("When I call remove duplicate and empty option function", func() {
			withoutEmpty := api.RemoveDuplicateAndEmptyOptions(duplicates)

			Convey("Then the empty values are removed", func() {
				expected := []string{"1", "2", "3"}
				So(withoutEmpty, ShouldResemble, expected)
			})
		})
	})
}

func TestRequestForwardingMiddleware(t *testing.T) {
	Convey("Given the assert dataset feature flag is toggled on", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true

		w := httptest.NewRecorder()

		Convey("When a POST request is made to the filters endpoint and the dataset type is cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "cantabular_flexible_table",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, datasetAPIMock, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is made to check the dataset type", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters endpoint and the dataset type is not cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "other",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, datasetAPIMock, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "other"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is made to check the dataset type", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When there is a PUT request to /filter-outputs/test-output-id and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check the filter output type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
			})
			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})
		Convey("When there is a PUT request to /filter-outputs/test-output-id and the filter type is NOT flexible", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "other"} }`)
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "NOT-flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			datastoreMock.CreateFilterOutputFunc = func(ctx context.Context, filter *models.Filter) error { return nil }

			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check the filter type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
			})
			Convey("The request is NOT forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When there is a PUT request to /filters/{id} and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("PUT", cfg().Host+"/filters/test-output-id", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check its filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/{id}/dimensions/{dimension}/options/{option} endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/{id}/dimensions/{dimension}/options/{option} endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Dataset: &models.Dataset{
						Version: 1,
					},
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given the assert dataset feature flag is toggled off", t, func() {
		conf := cfg()
		conf.AssertDatasetType = false

		w := httptest.NewRecorder()

		Convey("When a POST request is made to the filters endpoint and the dataset type is cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "cantabular_flexible_table",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is not made", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PUT request is made to the filter-outputs/id and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is NOT made to datastore to check the filter output type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("The request is NOT forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name/options/{option} endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: http.StatusOK,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, downloadSVCURL, externalDownloadSVCURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})
	})
}
