package api

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-filter-api/api/datastoretest"
	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	errAudit = errors.New("auditing error")
)

func TestSuccessfulAddFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	Convey("Given a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()

		mockDatastore := &datastoretest.DataStoreMock{
			AddFilterFunc: func(filter *models.Filter) (*models.Filter, error) {
				return filter, nil
			},
			CreateFilterOutputFunc: func(filter *models.Filter) error {
				return nil
			},
		}

		api := routes(host, mux.NewRouter(), mockDatastore, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filters endpoint", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)
			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionSuccessful, nil)
			})

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionSuccessful, nil)
			})

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionSuccessful, nil)
			})

			Convey("Then the data store is called to create a new filter output", func() {

				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, eventFilterOutputCreated)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
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

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filters endpoint", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r := createAuthenticatedRequest("POST", "http://localhost:22100/filters", reader)
			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionSuccessful, nil)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedToAddFilterBlueprint_AuditFailure(t *testing.T) {

	Convey("Given a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filters endpoint and the attempt audit fails", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], createFilterBlueprintAction, actionAttempted, nil)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint and the outcome audit fails", func() {

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == createFilterBlueprintAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionSuccessful, nil)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedToAddFilterBlueprint(t *testing.T) {
	t.Parallel()

	Convey("When duplicate dimensions are sent then a bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"},"dimensions":[{"name":"time","options":["Jun-15","Jun-12"]},{"name":"time","options":["Jun-14"]}]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldContainSubstring, "Bad request - duplicate dimension found: time")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no data store is available, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When dataset API is unavailable, an internal error is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{InternalServerError: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version does not exist, a not found error is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{VersionNotFound: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, versionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version does not exist, and auditing fails a 500 error is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			if action == createFilterBlueprintAction && result == actionUnsuccessful {
				return errAudit
			}
			return nil
		}

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{VersionNotFound: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version is unpublished and the request is not authenticated, a not found error is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, versionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
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

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filters endpoint which has an invalid JSON message", func() {

			reader := strings.NewReader("{")
			r, err := http.NewRequest("POST", "http://localhost:22100/filters", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
			})

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
			})

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
			})

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
			})

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimensions chosen: [weight]\n")
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

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, createFilterBlueprintAction, actionUnsuccessful, nil)
			})

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimension options chosen: [29]\n")
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

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}

	Convey("Given a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filters endpoint with no authentication", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 200 ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When a GET request is made to the filters endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getFilterBlueprintAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filters endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getFilterBlueprintAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_UnpublishedDataset(t *testing.T) {

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}
	t.Parallel()

	Convey("Successfully get an unpublished filter blueprint with authentication", t, func() {
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678", nil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprint(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a filter blueprint does not exist, and auditing fails a 500 error is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
			if action == getFilterBlueprintAction && result == actionUnsuccessful {
				return errAudit
			}
			return nil
		}

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished, and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678", nil)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestSuccessfulUpdateFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "21312"}

	Convey("Given a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()

		mockDatastore := &datastoretest.DataStoreMock{
			CreateFilterOutputFunc: func(filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(filterID string) (*models.Filter, error) {
				return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}, {Name: "1_age"}}}, nil
			},
			UpdateFilterFunc: func(filter *models.Filter, timestamp bson.MongoTimestamp) error {
				return nil
			},
		}

		api := routes(host, mux.NewRouter(), mockDatastore, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a PUT request is made to the filters endpoint", func() {

			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
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
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
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
			So(err, ShouldBeNil)

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the data store is called to create a new filter output", func() {

				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, eventFilterOutputCreated)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestFailedToUpdateFilterBlueprint_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "21312"}

	Convey("Given a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a PUT request is made to the filters endpoint and the attempt audit fails", func() {

			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], updateFilterBlueprintAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint and the outcome audit fails", func() {

			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == updateFilterBlueprintAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint with invalid json and the outcome audit fails", func() {

			reader := strings.NewReader("{")
			r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == updateFilterBlueprintAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
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

	expectedAuditParams := common.Params{"filter_blueprint_id": "21312"}

	Convey("Successfully send a request to submit an unpublished filter blueprint", t, func() {
		reader := strings.NewReader("{}")
		r := createAuthenticatedRequest("PUT", "http://localhost:22100/filters/21312?submitted=true", reader)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionSuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterBlueprint(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "21312"}

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When an empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to update filter blueprint that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no authentication is provided to update an unpublished filter, a not found is returned", t, func() {
		reader := strings.NewReader(`{"dimensions":[]}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()

		mockAuditor := getMockAuditor()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the version does not exist, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{VersionNotFound: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, versionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the datset version of a filter blueprint and the current dimensions do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [time 1_age]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the current dimension options do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", "http://localhost:22100/filters/21312", reader)
		So(err, ShouldBeNil)

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InvalidDimensionOption: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [28]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, updateFilterBlueprintAction, actionUnsuccessful, expectedAuditParams)
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestRemoveDuplicates(t *testing.T) {
	Convey("Given a string array with duplicate options", t, func() {
		duplicates := []string{"1", "2", "2", "2", "abcde", "abd", "abcde"}

		Convey("When I call remove duplicates function", func() {
			withoutDuplicates := removeDuplicateOptions(duplicates)

			Convey("Then the duplicates are removed", func() {
				expected := []string{"1", "2", "abcde", "abd"}
				So(withoutDuplicates, ShouldResemble, expected)

			})
		})
	})
}

func verifyAuditRecordCalls(c struct {
	Ctx    context.Context
	Action string
	Result string
	Params common.Params
}, expectedAction string, expectedResult string, expectedParams common.Params) {
	So(c.Action, ShouldEqual, expectedAction)
	So(c.Result, ShouldEqual, expectedResult)
	So(c.Params, ShouldResemble, expectedParams)
}

func assertAuditCalled(mockAuditor *audit.AuditorServiceMock, expectedAction, expectedOutcome string, expectedParams common.Params) {
	recCalls := mockAuditor.RecordCalls()
	So(len(recCalls), ShouldEqual, 2)
	verifyAuditRecordCalls(recCalls[0], expectedAction, actionAttempted, expectedParams)
	verifyAuditRecordCalls(recCalls[1], expectedAction, expectedOutcome, expectedParams)
}
