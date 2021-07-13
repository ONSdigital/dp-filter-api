package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-filter-api/filters"
)

func TestSuccessfulAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Given that a dimension option is successfully added to a filter", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datastoreMock := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), datastoreMock, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 201 Created status code is returned", func() {
			So(w.Code, ShouldEqual, http.StatusCreated)
		})

		Convey("Then the updated ETag is returned in a header", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag1)
		})

		Convey("The filter was requested to the data ase before and after being returned by the handler", func() {
			So(datastoreMock.GetFilterCalls(), ShouldHaveLength, 2)
			So(datastoreMock.GetFilterCalls()[0].ETagSelector, ShouldEqual, mongo.AnyETag)
			So(datastoreMock.GetFilterCalls()[1].ETagSelector, ShouldEqual, mongo.AnyETag)
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"33"})
		})
	})

	Convey("Given that a dimension option is successfully added to an unpublished filter", t, func() {
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Unpublished().Mock
		datastoreMock := mock.NewDataStore().Unpublished().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), datastoreMock, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 201 Created status code is returned", func() {
			So(w.Code, ShouldEqual, http.StatusCreated)
		})

		Convey("Then the updated ETag is returned in a header", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag1)
		})

		Convey("The filter was requested to the database before and after being returned by the handler", func() {
			So(datastoreMock.GetFilterCalls(), ShouldHaveLength, 2)
			So(datastoreMock.GetFilterCalls()[0].ETagSelector, ShouldEqual, mongo.AnyETag)
			So(datastoreMock.GetFilterCalls()[1].ETagSelector, ShouldEqual, mongo.AnyETag)
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"33"})
		})
	})
}

func TestFailedToAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Given that no data store is available, when trying to add a dimension option to a filter", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 500 InternalServerError status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Given that the filter blueprint does not exist, when trying to add a dimension option to a filter", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 400 BadRequest status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldResemble, filterNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Given that a filter blueprint is unpublished and the request is unauthenticated, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Unpublished().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 400 BadRequest status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldResemble, filterNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Given that a dimension option for filter blueprint does not exist, a bad request status is returned", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/66", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 400 BadRequest status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldResemble, "incorrect dimension options chosen: [66]\n")
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"66"})
		})
	})

	Convey("Given that a dimension for filter blueprint does not exist for that filter", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/notage/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 400 BadRequest status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			response := w.Body.String()
			So(response, ShouldResemble, "dimension not found\n")
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Given that a filter document has been modified by an external source", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().ConflictRequest(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 409 Conflict status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusConflict)
			So(w.Body.String(), ShouldContainSubstring, filters.ErrFilterBlueprintConflict.Error())
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Given that no If-Match header is provided", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		datasetAPIMock := mock.NewDatasetAPI().Mock
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().ConflictRequest(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)

		Convey("Then a 400 BadRequest status is returned with the expected error response", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, filters.ErrNoIfMatchHeader.Error())
		})

		Convey("Then the ETag header is empty", func() {
			So(w.HeaderMap.Get("ETag"), ShouldResemble, "")
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})
	})
}

func TestSuccessfulRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("Successfully remove a option for a filter blueprint, returns 204", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNoContent)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag1)
	})

	Convey("Successfully remove a option for an unpublished filter blueprint, returns 204", t, func() {
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, testETag1)
	})
}

func TestFailedToRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is not authenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().DimensionNotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})

	Convey("When the filter document has been modified by an external source, a conflict request status is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().ConflictRequest(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldContainSubstring, filters.ErrFilterBlueprintConflict.Error())
	})

	Convey("If no If-Match header is provided, then a 400 response is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldContainSubstring, filters.ErrNoIfMatchHeader.Error())
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()

	Convey("Given a mock returning a set of option dimensions", t, func() {

		expectedBodyFull := func() models.PublicDimensionOptions {
			return models.PublicDimensionOptions{
				Items: []*models.PublicDimensionOption{
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      &models.LinkObject{ID: "2014", HRef: "http://localhost:80/filters//dimensions/time/options/2014"},
							Filter:    &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						},
						Option: "2014",
					},
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      &models.LinkObject{ID: "2015", HRef: "http://localhost:80/filters//dimensions/time/options/2015"},
							Filter:    &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						},
						Option: "2015",
					},
				},
				Count:      2,
				Offset:     0,
				Limit:      20,
				TotalCount: 2,
			}
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
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
			validateBody(w.Body.Bytes(), expectedBodyFull())
		})

		Convey("Successfully get a list of dimension options for an unpublished filter blueprint", func() {
			r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
			validateBody(w.Body.Bytes(), expectedBodyFull())
		})

		Convey("Successfully get a list of dimensionOptions for a filter blueprint providing a zero offest", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?offset=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
			validateBody(w.Body.Bytes(), expectedBodyFull())
		})

		Convey("Successfully get the expected subset of dimensionOptions for a filter blueprint providing a non-zero offest", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?offset=1", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)

			expected := models.PublicDimensionOptions{
				Items: []*models.PublicDimensionOption{
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      &models.LinkObject{ID: "2015", HRef: "http://localhost:80/filters//dimensions/time/options/2015"},
							Filter:    &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						},
						Option: "2015",
					},
				},
				Count:      1,
				Offset:     1,
				Limit:      20,
				TotalCount: 2,
			}
			validateBody(w.Body.Bytes(), expected)
		})

		Convey("Successfully get the expected subset of dimensionOptions for a filter blueprint providing a non-zero limit", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=1&offset=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)

			expected := models.PublicDimensionOptions{
				Items: []*models.PublicDimensionOption{
					{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      &models.LinkObject{ID: "2014", HRef: "http://localhost:80/filters//dimensions/time/options/2014"},
							Filter:    &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Dimension: &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
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

		Convey("Successfully get the expected subset of dimensionOptions for a filter blueprint providing a limit greater than the total count", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=3&offset=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)

			expected := expectedBodyFull()
			expected.Limit = 3
			validateBody(w.Body.Bytes(), expected)
		})

		Convey("Successfully get dimensionOptions with empty list of items for a filter blueprint providing an offset greater than the total count", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?offset=3", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)

			expected := models.PublicDimensionOptions{
				Items:      []*models.PublicDimensionOption{},
				Count:      0,
				Offset:     3,
				Limit:      20,
				TotalCount: 2,
			}
			validateBody(w.Body.Bytes(), expected)
		})

		Convey("Successfully get dimensionOptions with empty list of items for a filter blueprint providing a zero limit value", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=0", nil)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
			filterApi.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)

			expected := models.PublicDimensionOptions{
				Items:      []*models.PublicDimensionOption{},
				Count:      0,
				Offset:     0,
				Limit:      0,
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
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When negative values are provided for limit and offset query parameters, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=-1&offset=-1", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When a limit higher than the maximum allowed is provided, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?limit=1001", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When an invalid offset value is provided, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options?offset=wrong", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a dimension not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().DimensionNotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

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
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
	})

	Convey("Successfully get a single dimension option for an unpublished filter blueprint", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
	})
}

func TestFailedToGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When option does not exist against filter blueprint, an option not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/age/options/notanage", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, optionNotFoundResponse)
	})
}

func TestSuccessfulPatchFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("Sending a single valid 'add' patch operation with 2 valid options, one already existing in the filter and one that does not exist", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27","33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag1)
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"27", "33"})
		})

		Convey("And only the valid inexistent option being added to the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 1)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.AddFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.AddFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"27"})
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Sending a single valid 'remove' patch operation with an existent option", t, func() {
		reader := strings.NewReader(`[
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag1)
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})

		Convey("And the option being removed from the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 1)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.RemoveFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"33"})
		})
	})

	Convey("Sending a single valid 'remove' patch operation with a mix of existent and inexistent options", t, func() {
		reader := strings.NewReader(`[
			{"op":"remove", "path": "/options/-", "value": ["33", "inexistent"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response ", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag1)
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})

		Convey("And only the existing options being updated to the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 1)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.RemoveFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"33"})
		})
	})

	Convey("Sending a single valid 'remove' patch operation with an inexistent option", t, func() {
		reader := strings.NewReader(`[
			{"op":"remove", "path": "/options/-", "value": ["inexistent"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})

		Convey("And no calls to remove options from the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 1)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Sending an empty list of patch operations", t, func() {
		reader := strings.NewReader(`[
			{"op":"remove", "path": "/options/-", "value": []}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag)
		})

		Convey("And no dimension or option is validated against DatasetAPI", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 0)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 0)
		})

		Convey("And no calls the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 0)
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 0)
		})
	})

	Convey("Sending a list of 2 valid patch operations", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27"]},
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag2)
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"27"})
		})

		Convey("And the expected calls for both operations are performed against the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 2)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.GetFilterCalls()[1].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.AddFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.AddFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"27"})
			So(ds.AddFilterDimensionOptionsCalls()[0].ETagSelector, ShouldResemble, testETag)
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.RemoveFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"33"})
			So(ds.RemoveFilterDimensionOptionsCalls()[0].ETagSelector, ShouldResemble, testETag1)
		})
	})

	Convey("Successfully patch dimension options for an unpublished filter blueprint", t, func() {
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27"]},
			{"op":"remove", "path": "/options/-", "value": ["33"]}
		]`)
		r := createAuthenticatedRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)

		ds := mock.NewDataStore().Unpublished().Mock
		datasetAPIMock := mock.NewDatasetAPI().Unpublished().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		Convey("Results in a 200 OK response, and the expected calls for both operations", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected ETag being returned", func() {
			So(w.HeaderMap.Get("ETag"), ShouldEqual, testETag2)
		})

		Convey("And the dimension and options are efficiently validated with dataset API", func() {
			So(datasetAPIMock.GetVersionDimensionsCalls(), ShouldHaveLength, 1)
			So(datasetAPIMock.GetOptionsBatchProcessCalls(), ShouldHaveLength, 1)
			So(*datasetAPIMock.GetOptionsBatchProcessCalls()[0].OptionIDs, ShouldResemble, []string{"27"})
		})

		Convey("And the expected calls for both operations are performed against the database", func() {
			So(ds.GetFilterCalls(), ShouldHaveLength, 2)
			So(ds.GetFilterCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.GetFilterCalls()[1].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.AddFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.AddFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.AddFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"27"})
			So(ds.RemoveFilterDimensionOptionsCalls(), ShouldHaveLength, 1)
			So(ds.RemoveFilterDimensionOptionsCalls()[0].FilterID, ShouldEqual, "12345678")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Name, ShouldEqual, "age")
			So(ds.RemoveFilterDimensionOptionsCalls()[0].Options, ShouldResemble, []string{"33"})
		})
	})
}

func TestFailedPatchBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("When a malformed patch is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`ASF$%£$^£@%$`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, "Failed to parse json body\n")
	})

	Convey("When a valid patch with an operation that is not supported is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`[{"op":"test", "path": "/options/-", "value": ["27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, "op 'test' not supported. Supported op(s): [add remove]\n")
	})

	Convey("When a valid patch with an incorrect path is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/wrong/path", "value": ["27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, "provided path '/wrong/path' not supported. Supported paths: '/options/-'\n")
	})

	Convey("Whe a valid 'add' patch with an incorrect option for a dimension is provided, a 400 BadRequest is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["wrong"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [wrong]\n")
	})

	Convey("When a valid patch with an overall sum of values higher than the maximum allowed is provided, a 400 BadRequest is returned", t, func() {
		cfg := cfg()
		cfg.MaxRequestOptions = 10
		reader := strings.NewReader(`[
			{"op":"add", "path": "/options/-", "value": ["27", "33", "27", "33", "27", "33", "27", "33"]},
			{"op":"add", "path": "/options/-", "value": ["27", "33", "27"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, "a maximum of 10 overall option values can be provied in a set of patch operations, which has been exceeded\n")
	})

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished())
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27", "33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().DimensionNotFound(), &mock.FilterJob{}, &mock.DatasetAPI{})
		filterApi.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.HeaderMap.Get("ETag"), ShouldEqual, "")

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})

	Convey("When the value of the provided If-Match header doesn't match the existing value in the database, a conflict response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27","33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", "oldValue")
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusConflict)
	})

	Convey("When no If-Match header is provided, a bad request response is returned", t, func() {
		reader := strings.NewReader(`[{"op":"add", "path": "/options/-", "value": ["27","33"]}]`)
		r, err := http.NewRequest("PATCH", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		ds := mock.NewDataStore().Mock
		datasetAPIMock := mock.NewDatasetAPI().Mock
		w := httptest.NewRecorder()
		filterApi := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, datasetAPIMock)
		filterApi.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}
