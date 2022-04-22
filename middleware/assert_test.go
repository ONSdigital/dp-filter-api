package middleware_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/middleware"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/ONSdigital/dp-filter-api/models"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-net/v2/responder"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterBlueprintID = "filter_blueprint_id"
	filterOutputID    = "filter_output_id"
	testBody          = "test body"
	testToken         = "testToken"
)

func setupFilterFlexAPIMock() *apimock.FilterFlexAPIMock {
	expectedResponse := &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(testBody))),
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"X-Test": []string{"Value"},
		},
	}

	return &apimock.FilterFlexAPIMock{
		ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
			return expectedResponse, nil
		},
	}
}

func setupDataStoreMock(filterIsFlexible bool, testBluePrintID string) *apimock.DataStoreMock {
	datastoreMock := mock.NewDataStore().Mock
	datastoreMock.GetFilterOutputFunc = func(ctx context.Context, filterOutputID string) (*models.Filter, error) {
		return &models.Filter{
			Dataset: &models.Dataset{
				Version: 1,
			},
			ID: filterOutputID,
			Links: models.LinkMap{
				FilterBlueprint: &models.LinkObject{ID: testBluePrintID, HRef: ""},
			},
		}, nil
	}
	datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
		filter := &models.Filter{
			Dataset: &models.Dataset{
				Version: 1,
			},
			ID: filterID,
		}
		if filterIsFlexible {
			filter.Type = "flexible"
		}
		return filter, nil
	}
	return datastoreMock
}

func setupDatasetAPIMock(datasetIsFlexible bool) *apimock.DatasetAPIMock {
	return &apimock.DatasetAPIMock{
		GetFunc: func(_ context.Context, _, _, _, _ string) (dataset.DatasetDetails, error) {
			dd := dataset.DatasetDetails{
				Type: "not_flexible",
			}
			if datasetIsFlexible {
				dd.Type = "cantabular_flexible_table"
			}
			return dd, nil
		},
	}
}

func assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled bool, testBluePrintID string) (*apimock.DataStoreMock, *apimock.FilterFlexAPIMock, *apimock.DatasetAPIMock, *middleware.Assert) {
	filterFlexAPIMock := setupFilterFlexAPIMock()
	datastoreMock := setupDataStoreMock(filterIsFlexible, testBluePrintID)
	datasetAPIMock := setupDatasetAPIMock(datasetIsFlexible)

	assert := middleware.NewAssert(
		responder.New(),
		datasetAPIMock,
		filterFlexAPIMock,
		datastoreMock,
		testToken,
		assertIsEnabled,
	)
	return datastoreMock, filterFlexAPIMock, datasetAPIMock, assert
}

func TestAssertFilterType(t *testing.T) {
	Convey("Given a healthy dp-cantabular-filter-flex-api", t, func() {
		testBluePrintID := "12345678"
		datasetIsFlexible := false
		Convey("When a filter with given ID in the datastore of type 'flexible' is found", func() {
			filterIsFlexible := true
			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assertIsEnabled := true
				datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testBluePrintID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as returned by dp-filter-flex-api", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterCalls()[0].FilterID, ShouldEqual, testBluePrintID)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, testBody)
				})
			})

			Convey("When an incoming request passes through a disabled assert.FilterType middleware", func() {
				assertIsEnabled := false
				datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testBluePrintID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})

		Convey("When a filter output with given ID in the datastore is found, but type is not 'flexible'", func() {
			assertIsEnabled := true
			filterIsFlexible := false
			datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

			Convey("When an incoming request passes through the assert.FilterOutputType middleware", func() {
				w := httptest.NewRecorder()
				testOutputID := "test-filter-output"
				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					"filter_output_id": testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputFilterType(next)
				f.ServeHTTP(w, r)

				Convey("The response should call the datastore with the same filter output ID", func() {
					So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterOutputCalls()[0].FilterOutputID, ShouldEqual, testOutputID)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
				Convey("The filter flex api request should not be forwarded.", func() {
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
				})

			})
		})

		Convey("When a filter with given ID in the datastore is found, but type is not 'flexible'", func() {
			filterIsFlexible := false
			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assertIsEnabled := true
				datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testBluePrintID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterCalls()[0].FilterID, ShouldEqual, testBluePrintID)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})
	})
}

func TestAssertFilterOutputType(t *testing.T) {
	Convey("Given a healthy dp-cantabular-filter-flex-api", t, func() {
		assertIsEnabled := true
		testBluePrintID := "filter-blueprint-123456"
		testOutputID := "12345678"
		datasetIsFlexible := false

		Convey("When a filter output with given filter in the datastore of type 'flexible' is found", func() {
			filterIsFlexible := true
			Convey("When an incoming request passes through the assert.FilterOutputType middleware", func() {
				datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterOutputID: testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as returned by dp-filter-flex-api", func() {
					So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterOutputCalls()[0].FilterOutputID, ShouldEqual, testOutputID)
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterCalls()[0].FilterID, ShouldEqual, testBluePrintID)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, testBody)
				})
			})

			Convey("When an incoming request passes through a disabled assert.FilterOutputType middleware", func() {
				assertIsEnabled = false
				datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterOutputID: testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})

		Convey("When a filter output with given ID in the datastore is found, but type is not 'flexible'", func() {
			filterIsFlexible := false
			datastoreMock, filterFlexAPIMock, _, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

			Convey("When an incoming request passes through the assert.FilterOutputType middleware", func() {
				w := httptest.NewRecorder()

				r, err := http.NewRequest("GET", "/some-url", nil)
				So(err, ShouldBeNil)

				requestWithVars := mux.SetURLVars(r, map[string]string{
					filterOutputID: testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
				f.ServeHTTP(w, requestWithVars)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterOutputCalls()[0].FilterOutputID, ShouldEqual, testOutputID)
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterCalls()[0].FilterID, ShouldEqual, testBluePrintID)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})
	})
}

func TestAssertDatasetType(t *testing.T) {
	Convey("Given a healthy dp-cantabular-filter-flex-api", t, func() {
		testBluePrintID := ""
		filterIsFlexible := true
		Convey("When a filter with given ID from dp-dataset-api of type 'cantabular_flexible_table' is returned", func() {
			datasetIsFlexible := true
			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assertIsEnabled := true
				_, filterFlexAPIMock, datasetAPIMock, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest(
					"POST",
					"http://localhost:1234/test",
					strings.NewReader(`{"dataset":{"version":1, "id":"cantabular-example-1"}}`),
				)
				So(err, ShouldBeNil)

				next := http.HandlerFunc(testHandler)
				f := assert.DatasetType(next)
				f.ServeHTTP(w, r)

				Convey("The response should have body, status and headers as returned by dp-filter-flex-api", func() {
					So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, testBody)
				})
			})

			Convey("When an incoming request passes through a disabled assert.FilterType middleware", func() {
				assertIsEnabled := false
				_, filterFlexAPIMock, datasetAPIMock, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)

				w := httptest.NewRecorder()

				r, err := http.NewRequest(
					"POST",
					"http://localhost:1234/test",
					strings.NewReader(`{"dataset":{"version":1, "id":"cantabular-example-1"}}`),
				)
				So(err, ShouldBeNil)

				next := http.HandlerFunc(testHandler)
				f := assert.DatasetType(next)
				f.ServeHTTP(w, r)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datasetAPIMock.GetCalls()), ShouldEqual, 0)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})

		Convey("When a filter with given ID from dp-dataset-api is found, but type is not 'cantabular_flexible_table'", func() {
			assertIsEnabled := true
			datasetIsFlexible := false
			_, filterFlexAPIMock, datasetAPIMock, assert := assertTestSetup(filterIsFlexible, datasetIsFlexible, assertIsEnabled, testBluePrintID)
			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				w := httptest.NewRecorder()

				r, err := http.NewRequest(
					"POST",
					"http://localhost:1234/test",
					strings.NewReader(`{"dataset":{"version":1, "id":"cantabular-example-1"}}`),
				)
				So(err, ShouldBeNil)

				next := http.HandlerFunc(testHandler)
				f := assert.DatasetType(next)
				f.ServeHTTP(w, r)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
					So(w.Code, ShouldEqual, http.StatusCreated)
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "Bar")
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, "test handler response")
				})
			})
		})
	})
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("X-Foo", "Bar")
	w.Write([]byte("test handler response"))
}
