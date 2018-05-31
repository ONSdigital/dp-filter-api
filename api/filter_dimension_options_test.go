package api

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/mocks"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessfulAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
		"option":              "33",
	}

	Convey("Successfully add a dimension option to a filter", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully add a dimension option to an unpublished filter", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToAddFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
		"option":              "33",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When the filter blueprint does not exist, a bad request status is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When the filter blueprint is unpublished, and the request is unauthenticated, a bad request status is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When the dimension option for filter blueprint does not exist, a bad request status is returned", t, func() {
		mockAuditor := getMockAuditor()
		expectedAuditParams := common.Params{
			"filter_blueprint_id": "12345678",
			"dimension":           "age",
			"option":              "66",
		}
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/66", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [66]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When a dimension for filter blueprint does not exist, a bad request status is returned", t, func() {
		mockAuditor := getMockAuditor()
		expectedAuditParams := common.Params{
			"filter_blueprint_id": "12345678",
			"dimension":           "notage",
			"option":              "33",
		}
		r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/notage/options/33", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InvalidDimensionOption: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [notage]\n")

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToAddFilterBlueprintDimensionOption_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "age",
		"option":              "33",
	}

	Convey("Given an existing published dataset and a filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filter options endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], addOptionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a POST request is made to the filter option endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == addOptionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, addOptionAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 201 Created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a POST request is made to the filter option endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("POST", "http://localhost:22100/filters/12345678/dimensions/age/options/33", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == addOptionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, addOptionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
		"option":              "2015",
	}

	Convey("Successfully remove a option for a filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully remove a option for an unpublished filter blueprint, returns 200", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToRemoveFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
		"option":              "26",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished, and request is not authenticated, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When dimension does not exist against filter blueprint, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, removeOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToRemoveFilterBlueprintDimensionOption_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
		"option":              "2015",
	}

	Convey("Given an existing published dataset and a filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a DELETE request is made to the filter options endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], removeOptionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a DELETE request is made to the filter option endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == removeOptionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, removeOptionAction, actionSuccessful, expectedAuditParams)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey("Given that the database returns an error when getting a filter output", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a DELETE request is made to the filter option endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("DELETE", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == removeOptionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, removeOptionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
	}

	Convey("Successfully get a list of dimension options for a filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get a list of dimension options for an unpublished filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensionOptions(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished and the request is unauthenticated, a not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When dimension does not exist against filter blueprint, a dimension not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{DimensionNotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, dimensionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionsAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensionOptions_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
	}

	Convey("Given an existing published dataset and a filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter options endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getOptionsAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter options endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getOptionsAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getOptionsAction, actionSuccessful, expectedAuditParams)
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

		Convey("When a GET request is made to the filter options endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getOptionsAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getOptionsAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
		"option":              "2015",
	}

	Convey("Successfully get a single dimension option for a filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionSuccessful, expectedAuditParams)
		})
	})

	Convey("Successfully get a single dimension option for an unpublished filter blueprint", t, func() {
		mockAuditor := getMockAuditor()
		r := createAuthenticatedRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNoContent)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionSuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensionOption(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "1_age",
		"option":              "26",
	}

	Convey("When no data store is available, an internal error is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InternalError: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint does not exist, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{NotFound: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When filter blueprint is unpublished, a bad request is returned", t, func() {
		mockAuditor := getMockAuditor()
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/1_age/options/26", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{Unpublished: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{Unpublished: true}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})

	Convey("When option does not exist against filter blueprint, an option not found is returned", t, func() {
		mockAuditor := getMockAuditor()
		expectedAuditParams := common.Params{
			"filter_blueprint_id": "12345678",
			"dimension":           "age",
			"option":              "notanage",
		}
		r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/age/options/notanage", nil)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{InvalidDimensionOption: true}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)

		response := w.Body.String()
		So(response, ShouldResemble, optionNotFoundResponse)

		Convey("Then the auditor is called for the attempt and outcome", func() {
			assertAuditCalled(mockAuditor, getOptionAction, actionUnsuccessful, expectedAuditParams)
		})
	})
}

func TestFailedToGetFilterBlueprintDimensionOption_AuditFailure(t *testing.T) {
	t.Parallel()

	expectedAuditParams := common.Params{
		"filter_blueprint_id": "12345678",
		"dimension":           "time",
		"option":              "2015",
	}

	Convey("Given an existing published dataset and a filter", t, func() {

		mockAuditor := getMockAuditor()
		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter(), &mocks.DataStore{}, &mocks.FilterJob{}, &mocks.DatasetAPI{}, previewMock, enablePrivateEndpoints, downloadServiceURL, downloadServiceToken, mockAuditor)

		Convey("When a GET request is made to the filter options endpoint and the attempt audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				return errAudit
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the action being attempted", func() {
				recCalls := mockAuditor.RecordCalls()
				So(len(recCalls), ShouldEqual, 1)
				verifyAuditRecordCalls(recCalls[0], getOptionAction, actionAttempted, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When a GET request is made to the filter option endpoint and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getOptionAction && result == actionSuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getOptionAction, actionSuccessful, expectedAuditParams)
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

		Convey("When a GET request is made to the filter option endpoint, and the outcome audit fails", func() {

			r, err := http.NewRequest("GET", "http://localhost:22100/filters/12345678/dimensions/time/options/2015", nil)
			So(err, ShouldBeNil)

			mockAuditor.RecordFunc = func(ctx context.Context, action string, result string, params common.Params) error {
				if action == getOptionAction && result == actionUnsuccessful {
					return errAudit
				}
				return nil
			}

			api.router.ServeHTTP(w, r)

			Convey("Then the auditor is called for the attempt and outcome", func() {
				assertAuditCalled(mockAuditor, getOptionAction, actionUnsuccessful, expectedAuditParams)
			})

			Convey("Then the response is 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
