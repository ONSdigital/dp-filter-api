package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

var host = "http://localhost:80"

func TestSuccessfulAddFilterJob(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		reader := strings.NewReader("{\"dataset_filter_id\":\"12345678\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestAddFilterFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader("{\"dataset_filter_id\":\"12345678\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a json message is missing mandatory fields, a bad request is returned", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestSuccessfulGetFilterJob(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a filter job", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestGetFilterFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter job does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestSuccessfulUpdateFilterJob(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\",\"version\":\"1\",\"edition\":\"1\",\"state\":\"submitted\"}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedUpdateFilterJob(t *testing.T) {
	t.Parallel()
	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a json message is missing mandatory fields, a bad request is returned", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\"}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a json message is sent to change a submitted filter on a filter job that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader("{\"state\":\"submitted\"}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("When a json message is sent to change a submitted filter, a status of forbidden is returned", t, func() {
		reader := strings.NewReader("{\"state\":\"submitted\"}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{Forbidden: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)
	})
}

func TestSuccessfulGetFilterDimensions(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimensions for a filter job", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestGetFilterDimensionsFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter job does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestSuccessfulGetFilterDimension(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a dimension for a filter job, returns 204", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestGetFilterDimensionFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter job does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When dimension does not exist against filter job, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestSuccessfulGetFilterDimensionOptions(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimension options for a filter job", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestGetFilterDimensionOptionsFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter job does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When dimension does not exist against filter job, a dimension not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestSuccessfulGetFilterDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("Successfully get a list of dimension options for a filter job", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestGetFilterDimensionOptionFailure(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/1234568/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("When filter job does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When option does not exist against filter job, an option not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{OptionNotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}
