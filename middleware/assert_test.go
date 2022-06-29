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
	testBody          = "test body"
	testToken         = "testToken"
)

func TestAssertFilterType(t *testing.T) {
	Convey("Given a healthy dp-cantabular-filter-flex-api", t, func() {
		expectedResponse := &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte(testBody))),
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"X-Test": []string{"Value"},
			},
		}

		filterFlexAPIMock := &apimock.FilterFlexAPIMock{
			ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
				return expectedResponse, nil
			},
		}

		Convey("When a filter with given ID in the datastore of type 'flexible' is found", func() {
			datastoreMock := mock.NewDataStore().Mock
			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
					ID: filterID,
				}, nil
			}

			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					true,
				)

				w := httptest.NewRecorder()

				testID := "12345678"
				r, err := http.NewRequest("GET", "http://localhost:1234/test/"+testID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, r)

				Convey("The response should have body, status and headers as returned by dp-filter-flex-api", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
					So(w.Code, ShouldEqual, expectedResponse.StatusCode)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, testBody)
				})
			})

			Convey("When an incoming request passes through a disabled assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					false,
				)

				w := httptest.NewRecorder()

				testID := "12345678"
				r, err := http.NewRequest("GET", "http://localhost:1234/test/"+testID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, r)

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
			datastoreMock := mock.NewDataStore().Mock
			datastoreMock.GetFilterOutputFunc = func(ctx context.Context, filterID string) (*models.Filter, error) {
				return &models.Filter{
					Type: "foobar",
					Dataset: &models.Dataset{
						Version: 2,
					},
					ID: filterID,
				}, nil
			}

			Convey("When an incoming request pases through the asert.FilterOutputType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					true,
				)

				w := httptest.NewRecorder()
				testOutputID := "test-filter-output"
				r, err := http.NewRequest("GET", "http://localhost:1234/filter-outputs/"+testOutputID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					"filter_output_id": testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
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

		Convey("When a filter output with given ID in the datastore is found, but type is 'flexible'", func() {
			datastoreMock := mock.NewDataStore().Mock
			datastoreMock.GetFilterOutputFunc = func(ctx context.Context, filterID string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 2,
					},
					ID: filterID,
				}, nil
			}

			Convey("When an incoming request pases through the asert.FilterOutputType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					true,
				)

				w := httptest.NewRecorder()
				testOutputID := "test-filter-output"
				r, err := http.NewRequest("GET", "http://localhost:1234/filter-outputs/"+testOutputID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					"filter_output_id": testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
				f.ServeHTTP(w, r)

				Convey("The response should call the datastore with the same filter output ID", func() {
					So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterOutputCalls()[0].FilterOutputID, ShouldEqual, testOutputID)
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.Body.String(), ShouldResemble, "test body")

				})

				Convey("The filter flex api request should be forwarded.", func() {
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
				})

			})

			Convey("When an incoming request pases through the assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					true,
				)

				w := httptest.NewRecorder()
				testOutputID := "test-filter-output"
				r, err := http.NewRequest("DELETE", "http://localhost:1234/filter-outputs/"+testOutputID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					"filter_output_id": testOutputID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterOutputType(next)
				f.ServeHTTP(w, r)

				Convey("The response should call the datastore with the same filter output ID", func() {
					So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
					So(datastoreMock.GetFilterOutputCalls()[0].FilterOutputID, ShouldEqual, testOutputID)
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.Body.String(), ShouldResemble, "test body")

				})

				Convey("The filter flex api request should be forwarded.", func() {
					So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
				})

			})
		})

		Convey("When a filter with given ID in the datastore is found, but type is not 'flexible'", func() {
			datastoreMock := mock.NewDataStore().Mock
			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "foobar",
					Dataset: &models.Dataset{
						Version: 2,
					},
					ID: filterID,
				}, nil
			}

			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					&mock.DatasetAPI{},
					filterFlexAPIMock,
					datastoreMock,
					testToken,
					true,
				)

				w := httptest.NewRecorder()

				testID := "12345678"
				r, err := http.NewRequest("GET", "http://localhost:1234/test/"+testID, nil)
				So(err, ShouldBeNil)

				r = mux.SetURLVars(r, map[string]string{
					filterBlueprintID: testID,
				})

				next := http.HandlerFunc(testHandler)
				f := assert.FilterType(next)
				f.ServeHTTP(w, r)

				Convey("The response should have body, status and headers as set by the 'next (testHandler)' function", func() {
					So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
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
		expectedResponse := &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte(testBody))),
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"X-Test": []string{"Value"},
			},
		}

		filterFlexAPIMock := &apimock.FilterFlexAPIMock{
			ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
				return expectedResponse, nil
			},
		}

		Convey("When a filter with given ID from dp-dataset-api of type 'cantabular_flexible_table' is returned", func() {
			datasetAPIMock := &apimock.DatasetAPIMock{}

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "cantabular_flexible_table",
				}, nil
			}

			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					datasetAPIMock,
					filterFlexAPIMock,
					&mock.DataStore{},
					testToken,
					true,
				)

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
					So(w.Code, ShouldEqual, expectedResponse.StatusCode)
					So(w.HeaderMap.Get("X-Test"), ShouldResemble, "Value")
					So(w.HeaderMap.Get("X-Foo"), ShouldResemble, "")
					So(w.Body.String(), ShouldResemble, testBody)
				})
			})

			Convey("When an incoming request passes through a disabled assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					datasetAPIMock,
					filterFlexAPIMock,
					&mock.DataStore{},
					testToken,
					false,
				)

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
			datasetAPIMock := &apimock.DatasetAPIMock{}

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "not_flexible",
				}, nil
			}

			Convey("When an incoming request passes through the assert.FilterType middleware", func() {
				assert := middleware.NewAssert(
					responder.New(),
					datasetAPIMock,
					filterFlexAPIMock,
					&mock.DataStore{},
					testToken,
					true,
				)

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
