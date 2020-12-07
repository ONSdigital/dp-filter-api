package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"io"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
)

func TestSuccessfulGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	Convey("Successfully get a list of dimensions for a filter blueprint", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Successfully get a list of dimensions for an unpublished filter blueprint", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailedToGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().InternalError(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().NotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})
}

func TestSuccessfulAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("Successfully create a dimension with an empty request body", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with a request body but no options", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with options", t, func() {
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Successfully create a dimension with options for an unpublished filter", t, func() {
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().InternalError(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

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
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

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
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().NotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

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
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

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
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

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
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [22]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When the filter document has been modified by an external source, a conflict request status is returned", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().ConflictRequest(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterBlueprintConflict.Error()+"\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestSuccessfulGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("Successfully get a dimension for a filter blueprint, returns 200", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		So(w.Body.String(), ShouldContainSubstring, `{"self":{"id":"1_age","href":"http://localhost:80/filters/12345678/dimensions/1_age"}`)
		So(w.Body.String(), ShouldContainSubstring, `"options":{"href":"http://localhost:80/filters/12345678/dimensions/1_age/options"`)
	})

	Convey("Successfully get a dimension for an unpublished filter blueprint, returns 200", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		So(w.Body.String(), ShouldContainSubstring, `{"self":{"id":"1_age","href":"http://localhost:80/filters/12345678/dimensions/1_age"}`)
		So(w.Body.String(), ShouldContainSubstring, `"options":{"href":"http://localhost:80/filters/12345678/dimensions/1_age/options"`)
	})
}

func TestFailedToGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().InternalError(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().NotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().DimensionNotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)
	})
}

func TestSuccessfulRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("Successfully remove a dimension for a filter blueprint, returns 204", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})

	Convey("Successfully remove a dimension for an unpublished filter blueprint, returns 204", t, func() {
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)
	})
}

func TestFailedToRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().InternalError(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().NotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished, and request is not authenticated, a bad request is returned", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().Unpublished(), &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When dimension does not exist against filter blueprint, the response is 404 Status Not Found", t, func() {
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := Setup(host, maxRequestOptions, mux.NewRouter(), mocks.NewDataStore().DimensionNotFound(), &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, "")
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
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

		publicDim := createPublicDimension(*testDim, "", "1234")

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

		publicDims := createPublicDimensions(testDims, "", "5678")

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
