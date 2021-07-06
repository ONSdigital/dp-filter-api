package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errAudit  = errors.New("auditing error")
	testETag  = fmt.Sprintf("%s0", mock.TestETag)
	testETag1 = fmt.Sprintf("%s1", mock.TestETag)
	testETag2 = fmt.Sprintf("%s2", mock.TestETag)
)

func TestSuccessfulAddFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	Convey("Given a published dataset", t, func() {

		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			AddFilterFunc: func(filter *models.Filter) (*models.Filter, error) {
				filter.ETag = testETag
				return filter, nil
			},
			CreateFilterOutputFunc: func(filter *models.Filter) error {
				return nil
			},
		}

		api := Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{})

		Convey("When a POST request is made to the filters endpoint", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)
			api.router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with valid dimensions", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)
			api.router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with the submitted query string parameter", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters?submitted=true", reader)
			So(err, ShouldBeNil)
			api.router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {

				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, eventFilterOutputCreated)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulAddFilterBlueprint_UnpublishedDataset(t *testing.T) {

	Convey("Given an unpublished dataset", t, func() {

		ds := mock.NewDataStore().Unpublished().Mock
		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())

		Convey("When a POST request is made to the filters endpoint", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filters", reader)
			api.router.ServeHTTP(w, r)

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag1)
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
	t.Parallel()

	Convey("When duplicate dimensions are sent", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"},"dimensions":[{"name":"time","options":["Jun-15","Jun-12"]},{"name":"time","options":["Jun-14"]}]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)

		Convey("Then the response is 400 bad request, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			response := w.Body.String()
			So(response, ShouldContainSubstring, "Bad request - duplicate dimension found: time")
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no data store is available", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When dataset API is unavailable", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().InternalServiceError())
		api.router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version does not exist", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound())
		api.router.ServeHTTP(w, r)

		Convey("Then the response is 404 Not Found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version is unpublished and the request is not authenticated", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		api.router.ServeHTTP(w, r)

		Convey("Then the response is 404 not found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToAddFilterBlueprint_BadJSON(t *testing.T) {

	Convey("Given a published dataset", t, func() {

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})

		Convey("When a POST request is made to the filters endpoint which has an invalid JSON message", func() {

			reader := strings.NewReader("{")
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which has an empty JSON message", func() {

			reader := strings.NewReader("{}")
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which is missing mandatory fields", func() {

			reader := strings.NewReader(`{"dataset":"Census"}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension that does not exist", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "weight", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimensions chosen: [weight]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension option that does not exist", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "age", "options": ["29","33"]}]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimension options chosen: [29]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
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
	t.Parallel()

	Convey("Given a published dataset", t, func() {

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})

		Convey("When a GET request is made to the filters endpoint with no authentication", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the response is 200 ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_UnpublishedDataset(t *testing.T) {
	t.Parallel()

	Convey("Successfully get an unpublished filter blueprint with authentication", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678", nil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the expected ETag is returned in the ETag header", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag)
		})
	})
}

func TestFailedToGetFilterBlueprint(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished, and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})
}

func TestSuccessfulUpdateFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	Convey("Given a published dataset", t, func() {
		testETag := "testETag"
		testETagUpdated := "testETagUpdated"
		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			CreateFilterOutputFunc: func(filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(filterID string, eTagSelector string) (*models.Filter, error) {
				if eTagSelector != mongo.AnyETag && eTagSelector != testETag {
					return nil, filters.ErrFilterBlueprintConflict
				}
				return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}, {Name: "1_age"}}, ETag: testETag}, nil
			},
			UpdateFilterFunc: func(filter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
				if eTagSelector != testETag {
					return "", filters.ErrFilterBlueprintConflict
				}
				return testETagUpdated, nil
			},
		}

		api := Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{})

		Convey("When a PUT request is made to the filters endpoint and a valid ETag", func() {

			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETagUpdated)
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
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETagUpdated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint to submit a filter blueprint", func() {

			reader := strings.NewReader("{}")
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312?submitted=true", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {

				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, eventFilterOutputCreated)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.HeaderMap.Get("ETag"), ShouldResemble, testETagUpdated)
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
	t.Parallel()

	Convey("Successfully send a request to submit an unpublished filter blueprint", t, func() {
		reader := strings.NewReader("{}")
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filters/21312?submitted=true", reader)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterBlueprint(t *testing.T) {
	t.Parallel()

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

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
			withoutDuplicates := removeDuplicateAndEmptyOptions(duplicates)

			Convey("Then the duplicates are removed", func() {
				expected := []string{"1", "2", "abcde", "abd"}
				So(withoutDuplicates, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a string array with empty options", t, func() {
		duplicates := []string{"", "1", "", "2", "", "3"}

		Convey("When I call remove duplicate and empty option function", func() {
			withoutEmpty := removeDuplicateAndEmptyOptions(duplicates)

			Convey("Then the empty values are removed", func() {
				expected := []string{"1", "2", "3"}
				So(withoutEmpty, ShouldResemble, expected)
			})
		})
	})
}
