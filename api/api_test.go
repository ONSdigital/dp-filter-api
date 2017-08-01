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
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{InternalError: true})
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
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("When a json message is missing mandatory fields, a bad request is returned", t, func() {
		reader := strings.NewReader("{\"dataset\":\"Census\"}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{BadRequest: true})
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
		api := CreateFilterAPI(host, mux.NewRouter(), &mocks.DataStore{})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}
