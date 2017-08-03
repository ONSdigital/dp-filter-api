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

func TestAddFilterJobReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\",\"version\":\"1\",\"edition\":\"1\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func TestAddFilterJobReturnsBadClientRequest(t *testing.T) {
	t.Parallel()
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

func TestSuccessfulAddFilterJobResponse(t *testing.T) {
	t.Parallel()
	Convey("Successfully send a valid json message", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\",\"version\":\"1\",\"edition\":\"1\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestSuccessfulUpdateFilterJobResponse(t *testing.T) {
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

func TestUpdateFilterJobReturnsBadClientRequest(t *testing.T) {
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
}

func TestUpdateFilterJobReturnsNotFoundClientRequest(t *testing.T) {
	t.Parallel()
	Convey("When a json message is sent to change a submitted filter on a filter job that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader("{\"state\":\"submitted\"}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestUpdateFilterJobReturnsForbiddenClientRequest(t *testing.T) {
	t.Parallel()
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
