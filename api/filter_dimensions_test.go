package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-filter-api/api"
	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"io"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
)

func TestSuccessfulGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given expected body", t, func() {
		expectedBodyFull := models.PublicDimensions{
			Items: []*models.PublicDimension{

				{
					Name: "1_age",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "1_age", HRef: "http://localhost:80/filters//dimensions/1_age"},
						Filter:  &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "http://localhost:80/filters//dimensions/1_age/options"},
					},
				},
				{
					Name: "age",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "age", HRef: "http://localhost:80/filters//dimensions/age"},
						Filter:  &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "http://localhost:80/filters//dimensions/age/options"},
					},
				},
				{
					Name: "time",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
						Filter:  &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "http://localhost:80/filters//dimensions/time/options"},
					},
				},
			},

			Count:      3,
			Offset:     0,
			Limit:      20,
			TotalCount: 3,
		}

		expectedBodyFullURLRewritten := models.PublicDimensions{
			Items: []*models.PublicDimension{
				{
					Name: "1_age",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "1_age", HRef: "https://api.test.com/filters/dimensions/1_age"},
						Filter:  &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/dimensions/1_age/options"},
					},
				},
				{
					Name: "age",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "age", HRef: "https://api.test.com/filters/dimensions/age"},
						Filter:  &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/dimensions/age/options"},
					},
				},
				{
					Name: "time",
					Links: &models.PublicDimensionLinkMap{
						Self:    &models.LinkObject{ID: "time", HRef: "https://api.test.com/filters/dimensions/time"},
						Filter:  &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/"},
						Options: &models.LinkObject{ID: "", HRef: "https://api.test.com/filters/dimensions/time/options"},
					},
				},
			},
			Count:      3,
			Offset:     0,
			Limit:      20,
			TotalCount: 3,
		}

		// func to unmarshal and validate body
		validateBody := func(bytes []byte, expected models.PublicDimensions) {
			var response models.PublicDimensions
			err := json.Unmarshal(bytes, &response)
			So(err, ShouldBeNil)
			So(response, ShouldResemble, expected)
		}

		Convey("Getting a list of dimensions for a filter blueprint results when url rewriting is enabled with authentication", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
			r.Header.Set("X-Forwarded-Host", "api.test.com")
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, true)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			validateBody(w.Body.Bytes(), expectedBodyFullURLRewritten)
		})

		Convey("Getting a list of dimensions for a filter blueprint results in a 200 response and expected body", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

			validateBody(w.Body.Bytes(), expectedBodyFull)
		})

		Convey("Getting a list of dimensions for an unpublished filter blueprint results in a 200 response and expected body", func() {
			r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

			validateBody(w.Body.Bytes(), expectedBodyFull)
		})

		Convey("Geting a list of dimensions with 0 offset results in a 200 response and expected body", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions?offset=0", http.NoBody)
			So(err, ShouldBeNil)

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

			validateBody(w.Body.Bytes(), expectedBodyFull)
		})

		Convey("Geting a list of dimensions with offset and limit results in a 200 response and expected items only", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions?offset=1&limit=3", http.NoBody)
			So(err, ShouldBeNil)

			expected := models.PublicDimensions{
				Items: []*models.PublicDimension{

					{
						Name: "age",
						Links: &models.PublicDimensionLinkMap{
							Self:    &models.LinkObject{ID: "age", HRef: "http://localhost:80/filters//dimensions/age"},
							Filter:  &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Options: &models.LinkObject{ID: "", HRef: "http://localhost:80/filters//dimensions/age/options"},
						},
					},
					{
						Name: "time",
						Links: &models.PublicDimensionLinkMap{
							Self:    &models.LinkObject{ID: "time", HRef: "http://localhost:80/filters//dimensions/time"},
							Filter:  &models.LinkObject{ID: "", HRef: "http://localhost:80/filters/"},
							Options: &models.LinkObject{ID: "", HRef: "http://localhost:80/filters//dimensions/time/options"},
						},
					},
				},

				Count:      2,
				Offset:     1,
				Limit:      3,
				TotalCount: 3,
			}

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

			validateBody(w.Body.Bytes(), expected)
		})

		Convey("Geting a list of dimensions with a zero limit results in a 200 response and an empty list of dimensions, with the correct totalCount", func() {
			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions?limit=0", http.NoBody)
			So(err, ShouldBeNil)

			expected := models.PublicDimensions{
				Items:      []*models.PublicDimension{},
				Count:      0,
				Offset:     0,
				Limit:      0,
				TotalCount: 3,
			}

			w := httptest.NewRecorder()
			filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
			filterAPI.Router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

			validateBody(w.Body.Bytes(), expected)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When negative values are provided for limit and offset query parameters, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions?offset=-8&limit=-3", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When a limit higher than the maximum allowed is provided, a bad request error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions?limit=1001", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, invalidQueryParameterResponse)
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})
}

func TestSuccessfulAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully create a dimension with an empty request body", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with a request body but no options", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with options", t, func() {
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with options for an unpublished filter", t, func() {
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
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

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
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

	Convey("When an unpublished filter blueprint does not exist, and the request is not authenticated, a not found is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
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

	Convey("When the dimension does not exist against the dataset filtered on, a bad request is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/wealth", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [wealth]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json body contains a dimension option that does not exist for a valid dimension, a bad request is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","33"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [22]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When the filter document has been modified by an external source before we obtain the filter, a conflict request status is returned", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().ConflictRequest(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterBlueprintConflict.Error()+"\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When the filter document has been modified by an external source after we obtained the filter and before we obtained the dimension, a conflict request status is returned", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		ds := mock.NewDataStore()
		ds.Mock.GetFilterDimensionFunc = func(ctx context.Context, filterID string, name, eTagSelector string) (dimension *models.Dimension, err error) {
			return nil, filters.ErrFilterBlueprintConflict
		}
		filterAPI := api.Setup(cfg(), mux.NewRouter(), ds.Mock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterBlueprintConflict.Error()+"\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When the If-Match header is not provided, a bad request status is returned", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
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

func TestSuccessfulGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully get a dimension for a filter blueprint, returns 200", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

		So(w.Body.String(), ShouldContainSubstring, `{"self":{"id":"1_age","href":"http://localhost:80/filters/12345678/dimensions/1_age"}`)
		So(w.Body.String(), ShouldContainSubstring, `"options":{"href":"http://localhost:80/filters/12345678/dimensions/1_age/options"`)
	})

	Convey("Successfully get a dimension for an unpublished filter blueprint, returns 200", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

		So(w.Body.String(), ShouldContainSubstring, `{"self":{"id":"1_age","href":"http://localhost:80/filters/12345678/dimensions/1_age"}`)
		So(w.Body.String(), ShouldContainSubstring, `"options":{"href":"http://localhost:80/filters/12345678/dimensions/1_age/options"`)
	})

	Convey("Getting a list of dimensions for a filter blueprint results when url rewriting is enabled with authentication", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("X-Forwarded-Host", "api.test.com")
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, true)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)

		So(w.Body.String(), ShouldContainSubstring, `{"self":{"id":"1_age","href":"https://api.test.com/filters/12345678/dimensions/1_age"}`)
		So(w.Body.String(), ShouldContainSubstring, `"options":{"href":"https://api.test.com/filters/12345678/dimensions/1_age/options"`)
	})
}

func TestFailedToGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().DimensionNotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})

	Convey("When an unpublished filter with a version that is published is updated by an external source after it is obtained from the datastore and before its state is updated, a conflict is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		ds := mock.NewDataStore().Unpublished().Mock
		ds.UpdateFilterFunc = func(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
			return "", filters.ErrFilterBlueprintConflict
		}
		filterAPI := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filerBlueprintConflictResponse)
	})
}

func TestSuccessfulRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Successfully remove a dimension for a filter blueprint, returns 204", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})

	Convey("Successfully remove a dimension for an unpublished filter blueprint, returns 204", t, func() {
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestFailedToRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished, and request is not authenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, the response is 404 Status Not Found", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().DimensionNotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Trying to remove an existing dimension without providing a valid If-Match header results in 400 Bad Request", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Trying to remove an existing dimension with an invalid value for If-Match header results in 409 Conflict", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", http.NoBody)
		r.Header.Set("If-Match", "wrong")
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)
	})
}

func TestCreatePublicDimensionSucceeds(t *testing.T) {
	t.Parallel()

	// Dimension test data
	testDim := &models.Dimension{
		URL:  "/filters/1234/dimensions/testDim1",
		Name: "testDim1",
	}

	Convey("When a Dimension struct is provided a PublicDimension struct is returned", t, func() {
		publicDim := api.CreatePublicDimension(*testDim, "", "1234")

		So(publicDim.Name, ShouldEqual, "testDim1")
		So(publicDim.Links.Self.ID, ShouldEqual, "testDim1")
		So(publicDim.Links.Self.HRef, ShouldEqual, testDim.URL)
		So(publicDim.Links.Filter.ID, ShouldEqual, "1234")
		So(publicDim.Links.Filter.HRef, ShouldEqual, "/filters/1234")
		So(publicDim.Links.Options.HRef, ShouldEqual, "/filters/1234/dimensions/testDim1/options")
	})
}

func TestCreatePublicDimensionsSucceeds(t *testing.T) {
	t.Parallel()

	// Dimensions test data
	testDims := []models.Dimension{
		{
			URL:  "/filters/5678/dimensions/testDim1",
			Name: "testDim1",
		},
		{
			URL:  "/filters/5678/dimensions/testDim2",
			Name: "testDim2",
		},
	}

	Convey("When an array of Dimension structs is provided an array of PublicDimension structs is returned", t, func() {
		publicDims := api.CreatePublicDimensions(testDims, "", "5678")

		So(len(publicDims), ShouldEqual, 2)

		So(publicDims[0].Name, ShouldEqual, "testDim1")
		So(publicDims[1].Name, ShouldEqual, "testDim2")
		So(publicDims[0].Links.Self.ID, ShouldEqual, "testDim1")
		So(publicDims[1].Links.Self.ID, ShouldEqual, "testDim2")
		So(publicDims[0].Links.Self.HRef, ShouldEqual, testDims[0].URL)
		So(publicDims[1].Links.Self.HRef, ShouldEqual, testDims[1].URL)
		So(publicDims[0].Links.Filter.ID, ShouldEqual, "5678")
		So(publicDims[1].Links.Filter.ID, ShouldEqual, "5678")
		So(publicDims[0].Links.Filter.HRef, ShouldEqual, "/filters/5678")
		So(publicDims[1].Links.Filter.HRef, ShouldEqual, "/filters/5678")
		So(publicDims[0].Links.Options.HRef, ShouldEqual, "/filters/5678/dimensions/testDim1/options")
		So(publicDims[1].Links.Options.HRef, ShouldEqual, "/filters/5678/dimensions/testDim2/options")
	})
}

func TestPutFilterBlueprintDimension(t *testing.T) {
	stubBody := `{
		"name":"Test dimension",
		"dimension_url":"http://localhost:22000/datasets/cantabular-flexible-example/editions/2021/versions/1/dimensions/city/options",
		"is_area_type":true
	}`
	reader := strings.NewReader(stubBody)

	Convey("Given the assert dataset feature flag is toggled off", t, func() {
		conf := cfg()
		conf.AssertDatasetType = false

		Convey("When the filter is flexible", func() {
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

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("An error is returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})
		})

		Convey("When the filter not flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("An error is returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given the assert dataset feature flag is toggled on", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true

		Convey("When the filter is flexible", func() {
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

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the dataset type", func() {
				So(datastoreMock.GetFilterCalls(), ShouldHaveLength, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 1)

				Convey("The request body is forwarded on", func() {
					call := filterFlexAPIMock.ForwardRequestCalls()[0]
					reqBody, err := io.ReadAll(call.Request.Body)
					So(err, ShouldBeNil)
					So(string(reqBody), ShouldEqual, stubBody)
				})
			})
		})

		Convey("When the filter not flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/12345678/dimensions/1_age", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("An error is returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})
		})
	})
}

func TestPostDimensionFilterFlexRoute(t *testing.T) {
	Convey("Given the assert dataset feature flag is toggled on", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true

		Convey("When the filter is flexible", func() {
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

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then a call to datastore is made to check the filter type", func() {
				So(datastoreMock.GetFilterCalls(), ShouldHaveLength, 1)
			})

			Convey("And the request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("When the filter not flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, mock.NewDatasetAPI().Mock, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})
		})
	})
}
