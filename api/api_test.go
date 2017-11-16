package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/gorilla/mux"

	"errors"
	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	host       = "http://localhost:80"
	authHeader = "cake"
	client     = rchttp.DefaultClient
	datasetAPI = mocks.NewDatasetAPI(client, "", "")
)

var previewMock = &datastoretest.PreviewDatasetMock{
	GetPreviewFunc: func(filter models.Filter) (*preview.FilterPreview, error) {
		return &preview.FilterPreview{}, nil
	},
}

func TestSuccessfulAddFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("Successfully create a filter blueprint", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})

	// TODO check test doesn't actually write job to queue?
	Convey("Successfully submit a filter blueprint", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters?submitted=true", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestFailedToAddFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When dataset API is unavailable, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{InternalServerError: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When instance does not exist, a not found error is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"12345678"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{InstanceNotFound: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Instance not found\n")
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When a json message is missing mandatory fields, a bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":"Census"}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})
}

func TestSuccessfulAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("Successfully create a dimension with an empty request body", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})

	Convey("Successfully create a dimension with a request body but no options", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})

	Convey("Successfully create a dimension with options", t, func() {
		reader := strings.NewReader(`{"values":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestFailedToAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"values":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		reader := strings.NewReader(`{"values":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})
}

func TestSuccessfulAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/65", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestFailedToAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/65", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When the filter blueprint does not exist, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/65", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When a dimension for filter blueprint does not exist, a not found status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/65", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})
}

func TestSuccessfulGetFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})
}

func TestSuccessfulUpdateFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		reader := strings.NewReader(`{"instance_id":"123"}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully send a request to submit filter blueprint", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312?submitted=true", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToUpdateFilterBlueprint(t *testing.T) {
	t.Parallel()
	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When an empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When a json message is sent to update filter blueprint that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"44444"}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})

	Convey("When a json message is sent to change the instance id of a filter blueprint and the instance does not exist, a status of not found is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"44444"}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InstanceNotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Instance not found\n")
	})
}

func TestSuccessfulGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimensions for a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})
}

func TestSuccessfulGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a dimension for a filter blueprint, returns 204", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestFailedToGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Dimension not found\n")
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimension options for a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When dimension does not exist against filter blueprint, a dimension not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Dimension not found\n")
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimension options for a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestFailedToGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When option does not exist against filter blueprint, an option not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{OptionNotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Option not found\n")
	})
}

func TestSuccessfulRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("Successfully remove a dimension for a filter blueprint, returns 200", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/1234568/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter blueprint not found\n")
	})
}

func TestSuccessfulRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("Successfully remove a option for a filter blueprint, returns 200", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request\n")
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Dimension not found\n")
	})
}

func TestSuccessfulGetFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a filter output", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
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

		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Failed to process the request due to an internal error\n")
	})

	Convey("When filter output does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filter-outputs/12345678", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Filter output not found\n")
	})
}

func TestSuccessfulUpdateFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToUpdateFilterOutput(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When an update to a filter output resource that does not exist, a not found is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})

	Convey("When a json message contains fields that are not allowed to be updated, a forbidden status is returned", t, func() {
		reader := strings.NewReader(`{"instance_id":"1234"}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cake")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Forbidden from updating the following fields: [instance_id]\n")
	})

	Convey("When a json message is sent to change a filter output with the wrong authorisation header, an unauthorised status is returned", t, func() {
		reader := strings.NewReader(`{"downloads":{"csv":{"url":"s3-csv-location","size":"12mb"}}}`)

		r, err := http.NewRequest("PUT", "http://localhost:22100/filter-outputs/21312", reader)
		So(err, ShouldBeNil)

		r.Header.Add(internalToken, "cookie")

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Unauthorised, request lacks valid authentication credentials\n")
	})
}

func TestSuccessfulGetPreview(t *testing.T) {
	t.Parallel()
	Convey("Successfully requesting a valid preview", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedGetPreview(t *testing.T) {
	t.Parallel()
	Convey("Requesting a preview with invalid filter", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Requesting a preview with no mongodb database connection", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Requesting a preview with no neo4j database connection", t, func() {
		previewMockInternalError := &datastoretest.PreviewDatasetMock{
			GetPreviewFunc: func(filter models.Filter) (*preview.FilterPreview, error) {
				return nil, errors.New("internal error")
			},
		}
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, datasetAPI, previewMockInternalError)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Requesting a preview with no dimensions", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/21312/preview", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(authHeader, host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{}, datasetAPI, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}
