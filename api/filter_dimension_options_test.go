package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-filter-api/filters"
)

func TestSuccessfulAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Successfully add a dimension option to a filter", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})

	Convey("Successfully add a dimension option to an unpublished filter", t, func() {
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestFailedToAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When the filter blueprint does not exist, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When the filter blueprint is unpublished, and the request is unauthenticated, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When the dimension option for filter blueprint does not exist, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/66", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [66]\n")
	})

	Convey("When a dimension for filter blueprint does not exist, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/notage/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InvalidDimensionOption: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "dimension not found\n")
	})

	Convey("When the filter document has been modified by an external source, a conflict request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{ConflictRequest: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)

		response := w.Body.String()
		So(response, ShouldContainSubstring, filters.ErrFilterBlueprintConflict.Error())
	})
}

func TestSuccessfulRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Successfully remove a option for a filter blueprint, returns 204", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})

	Convey("Successfully remove a option for an unpublished filter blueprint, returns 204", t, func() {
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestFailedToRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is not authenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})

	Convey("When the filter document has been modified by an external source, a conflict request status is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{ConflictRequest: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)

		response := w.Body.String()
		So(response, ShouldContainSubstring, filters.ErrFilterBlueprintConflict.Error())
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()

	Convey("Given a mock returning a set of option dimensions", t, func() {

		expectedBody := models.PublicDimensionOptions{
			Items: []*models.PublicDimensionOption{
				{
					Links: &models.PublicDimensionOptionLinkMap{
						Self:      models.LinkObject{ID: "2014", HRef: "http://localhost:80/filters//dimensions/time/options/2014"},
						Filter:    models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
						Dimension: models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
					},
					Option: "2014",
				},
				{
					Links: &models.PublicDimensionOptionLinkMap{
						Self:      models.LinkObject{ID: "2015", HRef: "http://localhost:80/filters//dimensions/time/options/2015"},
						Filter:    models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
						Dimension: models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
					},
					Option: "2015",
				},
			},
			Count:      2,
			Offset:     0,
			Limit:      0,
			TotalCount: 2,
		}

		// func to unmarshal and validate body
		validateBody := func(bytes []byte, expected models.PublicDimensionOptions) {
			var response models.PublicDimensionOptions
			err := json.Unmarshal(bytes, &response)
			So(err, ShouldBeNil)
			So(response, ShouldResemble, expected)
		}

		Convey("Successfully get a list of dimension options for a filter blueprint", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
			api.router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			validateBody(w.Body.Bytes(), expectedBody)
		})

		Convey("Successfully get a list of dimension options for an unpublished filter blueprint", func() {
			r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)

			w := httptest.NewRecorder()
			api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
			api.router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			validateBody(w.Body.Bytes(), expectedBody)
		})

		Convey("Successfully get a list of dimensionOptions for a filter blueprint providing values offest and limit", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=0&offset=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
			api.router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			validateBody(w.Body.Bytes(), expectedBody)
		})

		Convey("Successfully get the expected subset of dimensionOptions for a filter blueprint providing a non-zero offest", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=0&offset=1", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
			api.router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)

			expected := models.PublicDimensionOptions{
				Items: []*models.PublicDimensionOption{
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      models.LinkObject{ID: "2015", HRef: "http://localhost:80/filters//dimensions/time/options/2015"},
							Filter:    models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						},
						Option: "2015",
					},
				},
				Count:      1,
				Offset:     1,
				Limit:      0,
				TotalCount: 2,
			}
			validateBody(w.Body.Bytes(), expected)
		})

		Convey("Successfully get the expected subset of dimensionOptions for a filter blueprint providing a non-zero limit", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=1&offset=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
			api.router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)

			expected := models.PublicDimensionOptions{
				Items: []*models.PublicDimensionOption{
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      models.LinkObject{ID: "2014", HRef: "http://localhost:80/filters//dimensions/time/options/2014"},
							Filter:    models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						},
						Option: "2014",
					},
				},
				Count:      1,
				Offset:     0,
				Limit:      1,
				TotalCount: 2,
			}
			validateBody(w.Body.Bytes(), expected)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()

	Convey("When an invalid limit value is provided, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=wrong&offset=0", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When an invalid offset value is provided, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=0&offset=wrong", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a dimension not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Successfully get a single dimension option for a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully get a single dimension option for an unpublished filter blueprint", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When option does not exist against filter blueprint, an option not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/age/options/notanage", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InvalidDimensionOption: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, optionNotFoundResponse)
	})
}

func TestSuccessfulPatchFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("Successfully patch dimension options, with a single valid 'add' patch operation, returns 200", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27","33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully patch dimension options, with a single valid 'remove' patch operation, returns 200", t, func() {
		reader := strings.NewReader(`[
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully patch dimension options, with a combination of valid patch operations, returns 200", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27"]},
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully patch dimension options for an unpublished filter blueprint", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27"]},
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r := createAuthenticatedRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedPatchBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("When a malformed patch is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`ASF$%£$^£@%$`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "Failed to parse json body\n")
	})

	Convey("When a valid patch with an operation that is not supported is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`[{"op":"test", "path": "/options/-", "value": ["27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "op 'test' not supported. Supported op(s): [add remove]\n")
	})

	Convey("When a valid patch with an incorrect path is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/wrong/path", "value": ["27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "provided path '/wrong/path' not supported. Supported paths: '/options/-'\n")
	})

	Convey("When a valid patch with an overall sum of values higher than the maximum allowed is provided, a 400 BadRequest is returned", t, func() {
		cfg := cfg()
		cfg.MaxRequestOptions = 10
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27", "33", "27", "33", "27", "33", "27", "33"]},
			{"op":"add", "path": "/options/-", "value": ["27", "33", "27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "a maximum of 10 overall option values can be provied in a set of patch operations, which has been exceeded\n")
	})

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(cfg(), mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})
}
