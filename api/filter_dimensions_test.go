package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
)

func TestSuccessfulGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}

	Convey("Successfully get a list of dimensions for a filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionsAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get a list of dimensions for an unpublished filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionsAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensions(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensions_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{"filter_blueprint_id": "12345678"}

	Convey("Given an existing published filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter dimensions endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getDimensionsAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter dimensions endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getDimensionsAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getDimensionsAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter dimensions endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getDimensionsAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getDimensionsAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
	}

	Convey("Successfully create a dimension with an empty request body", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully create a dimension with a request body but no options", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully create a dimension with options", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully create a dimension with options for an unpublished filter", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["27","33"]}`)
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToAddFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, badRequestResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When an unpublished filter blueprint does not exist, and the request is not authenticated, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When the dimension does not exist against the dataset filtered on, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		expectedAuditParams := common.Params{
			"filter_blueprint_id": "12345678",
			"dimension":           "wealth",
		}
		reader := strings.NewReader(`{"options":["22","17"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/wealth", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [wealth]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a json body contains a dimension option that does not exist for a valid dimension, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		reader := strings.NewReader(`{"options":["22","33"]}`)
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [22]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToAddFilterBlueprintDimension_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
	}

	Convey("Given an existing filter for a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filter dimension endpoint and the attempt audit fails", func() {

			reader := strings.NewReader(`{"options":["27","33"]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], addDimensionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a POST request is made to the filter dimension endpoint and the outcome audit fails", func() {

			reader := strings.NewReader(`{"options":["27","33"]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == addDimensionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, addDimensionAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filter dimension endpoint, and the outcome audit fails", func() {

			reader := strings.NewReader(`{"options":["27","33"]}`)
			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == addDimensionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, addDimensionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("When the filter document has been modified by an external source, a conflict request status is returned", t, func() {
		reader := strings.NewReader("")
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{ConflictRequest: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, getMockAuditor())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusConflict)

		response := w.Body.String()
		So(response, ShouldResemble, filters.ErrFilterBlueprintConflict.Error()+"\n")
	})
}

func TestSuccessfulGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("Successfully get a dimension for a filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get a dimension for an unpublished filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished and request is unauthenticated, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimension_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("Given an existing published filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter dimension endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getDimensionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter dimension endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getDimensionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getDimensionAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter dimension endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getDimensionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getDimensionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("Successfully remove a dimension for a filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully remove a dimension for an unpublished filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToRemoveFilterBlueprintDimension(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished, and request is not authenticated, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When dimension does not exist against filter blueprint, the response is 404 Status Not Found", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeDimensionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToRemoveFilterBlueprintDimension_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("Given an existing filter for a published dataset", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a DELETE request is made to the filter dimension endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], removeDimensionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a DELETE request is made to the filter dimension endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == removeDimensionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, removeDimensionAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 200 ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a DELETE request is made to the filter dimension endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == removeDimensionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, removeDimensionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestCreatePublicDimensionSucceeds(t *testing.T) {
	t.Parallel()

	// Dimenson test data
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

	// Dimensons test data
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
