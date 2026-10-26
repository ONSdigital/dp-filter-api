package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filterflex"
	"github.com/ONSdigital/dp-filter-api/api"
	apimock "github.com/ONSdigital/dp-filter-api/api/mock"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/mock"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	dprequest "github.com/ONSdigital/dp-net/request"
	dprequestv2 "github.com/ONSdigital/dp-net/v2/request"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	testETag      = fmt.Sprintf("%s0", mock.TestETag)
	testETag1     = fmt.Sprintf("%s1", mock.TestETag)
	testETag2     = fmt.Sprintf("%s2", mock.TestETag)
	hostURL       = &url.URL{Scheme: "http", Host: "localhost:80"}
	datasetAPIURL = &url.URL{Scheme: "http", Host: "localhost:22000"}
)

func TestSuccessfulAddFilterBlueprint_PublishedDataset(t *testing.T) {
	t.Parallel()

	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			AddFilterFunc: func(ctx context.Context, filter *models.Filter) (*models.Filter, error) {
				filter.ETag = testETag
				return filter, nil
			},
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
		}

		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		Convey("When a POST request is made to the filters endpoint", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with valid dimensions", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with the submitted query string parameter", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters?submitted=true", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, models.EventFilterOutputCreated)
			})

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestSuccessfulPostFilterBlueprint_Flexible(t *testing.T) {
	Convey("Given an flexible filter", t, func() {
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
		filterAPI := api.Setup(cfg(), mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		Convey("When a post request is made to the submit endpoint", func() {
			reader := strings.NewReader(`{}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters/123456/submit", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the request is not be forwarded to the filter flex api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 0)
			})

			Convey("Then an error is returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestFlexibleFilterForwardsRequestWithoutActiveTracing(t *testing.T) {
	type receivedRequest struct {
		method string
		path   string
		query  url.Values
		header http.Header
		body   string
		err    error
	}

	received := make(chan receivedRequest, 1)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		received <- receivedRequest{r.Method, r.URL.Path, r.URL.Query(), r.Header.Clone(), string(body), err}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer backend.Close()

	conf := cfg()
	conf.AssertDatasetType = true
	conf.EnablePrivateEndpoints = false
	store := mock.NewDataStore().Mock
	store.GetFilterFunc = func(_ context.Context, _, _ string) (*models.Filter, error) {
		return &models.Filter{Type: "flexible"}, nil
	}

	router := mux.NewRouter()
	noTracing := noop.NewTracerProvider()
	router.Use(otelmux.Middleware("dp-filter-api", otelmux.WithTracerProvider(noTracing)))
	api.Setup(conf, router, store, &mock.FilterJob{}, &mock.DatasetAPI{},
		filterflex.New(filterflex.Config{HostURL: backend.URL}), hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
	handler := dprequest.HandlerRequestID(16)(dprequestv2.HandlerRequestID(16)(otelhttp.NewMiddleware("dp-filter-api", otelhttp.WithTracerProvider(noTracing))(router)))

	const body = `{"options":["27"]}`
	const traceparent = "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01"
	const requestID = "caller-request"
	request := httptest.NewRequest(http.MethodPut, "/filters/filter-123/dimensions/age?choice=blue&choice=green", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer synthetic-user")
	request.Header.Set("Cookie", "session=synthetic-session")
	request.Header.Set("X-Forwarded-Host", "api.example.test")
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("X-Request-Id", requestID)
	request.Header.Set("traceparent", traceparent)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("forwarded response status: got %d, want %d", response.Code, http.StatusAccepted)
	}

	select {
	case got := <-received:
		if got.err != nil {
			t.Fatal(got.err)
		}
		if got.method != http.MethodPut || got.path != "/filters/filter-123/dimensions/age" {
			t.Errorf("backend request: got %s %s", got.method, got.path)
		}
		if want := (url.Values{"choice": {"blue", "green"}}); !reflect.DeepEqual(got.query, want) {
			t.Errorf("backend query: got %v, want %v", got.query, want)
		}
		if got.body != body {
			t.Errorf("backend body: got %q, want %q", got.body, body)
		}
		for key, want := range map[string]string{
			"Content-Type":         "application/json",
			"Authorization":        "Bearer synthetic-user",
			"Cookie":               "session=synthetic-session",
			"X-Forwarded-Host":     "api.example.test",
			"X-Forwarded-Proto":    "https",
			"X-Forwarded-API-Host": "api.example.test",
			"traceparent":          traceparent,
		} {
			if value := got.header.Get(key); value != want {
				t.Errorf("backend %s: got %q, want %q", key, value, want)
			}
		}
		ids := got.header.Values("X-Request-Id")
		if len(ids) != 2 || ids[0] != requestID || !strings.HasPrefix(ids[1], requestID+",") || len(ids[1]) != len(requestID)+1+len(requestID)/2 {
			t.Errorf("backend X-Request-Id: got %q, want caller ID extended with a new hop", ids)
		}
	default:
		t.Fatal("backend received no forwarded request")
	}
}

func TestFlexibleFilterPreservesIncomingTraceWithOTel(t *testing.T) {
	t.Setenv("OTEL_PROPAGATORS", "tracecontext,baggage")
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})
	otel.SetTracerProvider(provider)
	propagator := autoprop.NewTextMapPropagator()
	otel.SetTextMapPropagator(propagator)

	received := make(chan http.Header, 1)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	conf := cfg()
	conf.AssertDatasetType = true
	conf.EnablePrivateEndpoints = false
	store := mock.NewDataStore().Mock
	store.GetFilterFunc = func(_ context.Context, _, _ string) (*models.Filter, error) {
		return &models.Filter{Type: "flexible"}, nil
	}
	router := mux.NewRouter()
	router.Use(otelmux.Middleware(conf.OTServiceName))
	api.Setup(conf, router, store, &mock.FilterJob{}, &mock.DatasetAPI{},
		filterflex.New(filterflex.Config{HostURL: backend.URL}), hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
	handler := dprequest.HandlerRequestID(16)(dprequestv2.HandlerRequestID(16)(otelhttp.NewMiddleware(conf.OTServiceName)(router)))

	const traceparent = "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01"
	const requestID = "caller-request"
	request := httptest.NewRequest(http.MethodGet, "/filters/filter-123", http.NoBody)
	request.Header.Set("traceparent", traceparent)
	request.Header.Set("X-Request-Id", requestID)
	incomingSpan := trace.SpanContextFromContext(propagator.Extract(context.Background(), propagation.HeaderCarrier(request.Header)))
	if !incomingSpan.IsValid() {
		t.Fatalf("invalid incoming traceparent: %q", traceparent)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("forwarded response status: got %d, want %d", response.Code, http.StatusOK)
	}

	select {
	case headers := <-received:
		backendSpan := trace.SpanContextFromContext(propagator.Extract(context.Background(), propagation.HeaderCarrier(headers)))
		if !backendSpan.IsValid() {
			t.Fatalf("backend received invalid traceparent: %q", headers.Get("traceparent"))
		}
		if backendSpan.TraceID() != incomingSpan.TraceID() || backendSpan.SpanID() != incomingSpan.SpanID() {
			t.Errorf("backend trace context: got %s/%s, want %s/%s", backendSpan.TraceID(), backendSpan.SpanID(), incomingSpan.TraceID(), incomingSpan.SpanID())
		}
		ids := headers.Values("X-Request-Id")
		if len(ids) != 2 || ids[0] != requestID || !strings.HasPrefix(ids[1], requestID+",") || len(ids[1]) != len(requestID)+1+len(requestID)/2 {
			t.Errorf("backend X-Request-Id: got %q, want caller ID extended with a new hop", ids)
		}
	default:
		t.Fatal("backend received no forwarded request")
	}

	activeSpan := false
	for _, span := range exporter.GetSpans() {
		if span.SpanContext.TraceID() == incomingSpan.TraceID() && span.SpanContext.SpanID() != incomingSpan.SpanID() {
			activeSpan = true
		}
	}
	if !activeSpan {
		t.Error("inbound OTel middleware did not record a child span of the incoming trace")
	}
}

func TestFlexibleFilterDatasetLookupUsesServiceRequest(t *testing.T) {
	type receivedRequest struct {
		method string
		path   string
		query  url.Values
		header http.Header
		body   string
		err    error
	}

	lookups := make(chan receivedRequest, 1)
	datasetBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		lookups <- receivedRequest{r.Method, r.URL.Path, r.URL.Query(), r.Header.Clone(), string(body), err}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"type":"cantabular_flexible_table"}`)
	}))
	defer datasetBackend.Close()

	forwarded := make(chan string, 1)
	flexBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		forwarded <- string(body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer flexBackend.Close()

	datasetURL, err := url.Parse(datasetBackend.URL)
	if err != nil {
		t.Fatal(err)
	}
	conf := cfg()
	conf.AssertDatasetType = true
	conf.EnablePrivateEndpoints = false
	conf.ServiceAuthToken = "synthetic-service"
	router := mux.NewRouter()
	noTracing := noop.NewTracerProvider()
	router.Use(otelmux.Middleware("dp-filter-api", otelmux.WithTracerProvider(noTracing)))
	api.Setup(conf, router, &mock.DataStore{}, &mock.FilterJob{}, dataset.NewAPIClient(datasetBackend.URL),
		filterflex.New(filterflex.Config{HostURL: flexBackend.URL}), hostURL, datasetURL, parsedDownloadServiceURL, false)
	handler := dprequest.HandlerRequestID(16)(dprequestv2.HandlerRequestID(16)(otelhttp.NewMiddleware("dp-filter-api", otelhttp.WithTracerProvider(noTracing))(router)))

	const body = `{"dataset":{"id":"dataset-123","edition":"2026","version":1}}`
	const requestID = "caller-request"
	request := httptest.NewRequest(http.MethodPost, "/filters?choice=blue&choice=green", strings.NewReader(body))
	request.Header.Set("X-Request-Id", requestID)
	request.Header.Set("Authorization", "Bearer synthetic-caller")
	request.Header.Set("Cookie", "session=synthetic-session")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("forwarded response status: got %d, want %d", response.Code, http.StatusAccepted)
	}

	select {
	case got := <-lookups:
		if got.err != nil {
			t.Fatal(got.err)
		}
		if got.method != http.MethodGet || got.path != "/datasets/dataset-123" {
			t.Errorf("dataset lookup: got %s %s", got.method, got.path)
		}
		if len(got.query) != 0 || got.body != "" {
			t.Errorf("dataset lookup should not forward POST query or body: got query %v, body %q", got.query, got.body)
		}
		if auth := got.header.Get("Authorization"); auth != "Bearer synthetic-service" {
			t.Errorf("dataset lookup Authorization: got %q, want service token", auth)
		}
		if cookie := got.header.Get("Cookie"); cookie != "" {
			t.Errorf("dataset lookup Cookie: got %q, want none", cookie)
		}
		if contentType := got.header.Get("Content-Type"); contentType != "" {
			t.Errorf("dataset lookup Content-Type: got %q, want none", contentType)
		}
		ids := got.header.Values("X-Request-Id")
		if len(ids) != 1 || !strings.HasPrefix(ids[0], requestID+",") {
			t.Errorf("dataset lookup X-Request-Id: got %q, want %q extended with a new hop", ids, requestID)
		}
	default:
		t.Fatal("dataset backend received no lookup")
	}

	select {
	case got := <-forwarded:
		if got != body {
			t.Errorf("Flex POST body: got %q, want %q", got, body)
		}
	default:
		t.Fatal("Flex backend received no POST")
	}
}

func TestUnSuccessfulPostFilterBlueprint_NonFlexible(t *testing.T) {
	Convey("Given an non flexible filter", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true
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

		Convey("When a post request is made to the submit endpoint", func() {
			reader := strings.NewReader(`{}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters/123456/submit", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(filterFlexAPIMock.ForwardRequestCalls(), ShouldHaveLength, 1)

				Convey("The request body is forwarded on", func() {
					call := filterFlexAPIMock.ForwardRequestCalls()[0]
					reqBody, err := io.ReadAll(call.Request.Body)
					So(err, ShouldBeNil)
					So(string(reqBody), ShouldEqual, `{}`)
				})
			})
		})
	})
}

func TestSuccessfulAddFilterBlueprint_UnpublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given an unpublished dataset", t, func() {
		ds := mock.NewDataStore().Unpublished().Mock
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), ds, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		Convey("When a POST request is made to the filters endpoint", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
			r := createAuthenticatedRequest("POST", cfg().Host+"/filters", reader)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 201 created", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Then the expected ETag is returned in a header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
				So(ds.AddFilterCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestFailedToAddFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When duplicate dimensions are sent", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"},"dimensions":[{"name":"time","options":["Jun-15","Jun-12"]},{"name":"time","options":["Jun-14"]}]}`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 400 bad request, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			response := w.Body.String()
			So(response, ShouldContainSubstring, "Bad request - duplicate dimension found: time")
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When no data store is available", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When dataset API is unavailable", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().InternalServiceError(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 500 internal error, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			response := w.Body.String()
			So(response, ShouldResemble, internalErrResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version does not exist", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} }`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 404 Not Found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When version is unpublished and the request is not authenticated", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"}, "dimensions":[{"name": "age", "options": ["27","33"]}]}`)
		r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)

		Convey("Then the response is 404 not found, with the expected response body", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			response := w.Body.String()
			So(response, ShouldResemble, versionNotFoundResponse)
		})

		Convey("Then the ETag header is empty", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, "")
		})

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToAddFilterBlueprint_BadJSON(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		Convey("When a POST request is made to the filters endpoint which has an invalid JSON message", func() {
			reader := strings.NewReader("{")
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which has an empty JSON message", func() {
			reader := strings.NewReader("{}")
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint which is missing mandatory fields", func() {
			reader := strings.NewReader(`{"dataset":"Census"}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, badRequestResponse)
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension that does not exist", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "weight", "options": ["27","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimensions chosen: [weight]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a POST request is made to the filters endpoint with a dimension option that does not exist", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1"} , "dimensions":[{"name": "age", "options": ["29","33"]}]}`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 400 bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Then the response body contains the expected content", func() {
				So(w.Body.String(), ShouldResemble, "incorrect dimension options chosen: [29]\n")
			})

			Convey("Then the ETag header is empty", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, "")
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
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Given a published dataset", t, func() {
		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, true)

		Convey("When a GET request is made to the filters endpoint with no authentication", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the response is 200 ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_PublishedDataset_URLRewriting(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}
	t.Parallel()

	Convey("Given a published dataset with URL rewriting enabled", t, func() {
		w := httptest.NewRecorder()
		mockDatastore := &apimock.DataStoreMock{
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
				if eTagSelector != mongo.AnyETag && eTagSelector != testETag {
					return nil, filters.ErrFilterBlueprintConflict
				}
				return &models.Filter{
					Dataset:    &models.Dataset{ID: "123", Edition: "2019", Version: 1},
					InstanceID: "12345678",
					Published:  &models.Published,
					Dimensions: []models.Dimension{
						{Name: "time", Options: []string{"2014", "2015"}},
						{Name: "1_age"},
					},
					ETag: testETag,
					Links: models.LinkMap{
						Self:       &models.LinkObject{HRef: fmt.Sprintf("%s/filters/%s", hostURL, filterID)},
						Dimensions: &models.LinkObject{HRef: fmt.Sprintf("%s/filters/%s/dimensions", hostURL, filterID)},
						Version:    &models.LinkObject{HRef: fmt.Sprintf("%s/datasets/123/editions/2019/versions/1", hostURL)},
					},
				}, nil
			},
		}
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, true)

		Convey("When a GET request is made to the filters endpoint without X-Forwarded-Host", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)
			bodyBytes := w.Body.Bytes()

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the rewritten links should be correct", func() {
				var responseBody map[string]interface{}
				err := json.Unmarshal(bodyBytes, &responseBody)
				So(err, ShouldBeNil)

				links, ok := responseBody["links"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				selfLink, ok := links["self"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(selfLink["href"], ShouldEqual, "http://localhost:80/filters/12345678")

				dimensionsLink, ok := links["dimensions"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(dimensionsLink["href"], ShouldEqual, "http://localhost:80/filters/12345678/dimensions")

				versionLink, ok := links["version"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(versionLink["href"], ShouldEqual, "http://localhost:22000/datasets/123/editions/2019/versions/1")
			})
		})

		Convey("When a GET request is made to the filters endpoint with X-Forwarded-Host", func() {
			r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
			r.Header.Set("X-Forwarded-Host", "api.test.com")
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)
			bodyBytes := w.Body.Bytes()

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the expected ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
			})

			Convey("Then the rewritten links should be correct", func() {
				var responseBody map[string]interface{}
				err := json.Unmarshal(bodyBytes, &responseBody)
				So(err, ShouldBeNil)
				links, ok := responseBody["links"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				selfLink, ok := links["self"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(selfLink["href"], ShouldEqual, "https://api.test.com/filters/12345678")

				dimensionsLink, ok := links["dimensions"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(dimensionsLink["href"], ShouldEqual, "https://api.test.com/filters/12345678/dimensions")

				versionLink, ok := links["version"].(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(versionLink["href"], ShouldEqual, "https://api.test.com/datasets/123/editions/2019/versions/1")
			})
		})
	})
}

func TestSuccessfulGetFilterBlueprint_UnpublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Successfully get an unpublished filter blueprint with authentication", t, func() {
		r := createAuthenticatedRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		Convey("Then the expected ETag is returned in the ETag header", func() {
			So(w.Result().Header.Get("ETag"), ShouldResemble, testETag)
		})
	})
}

func TestFailedToGetFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When no data store is available, an internal error is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InternalError(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, internalErrResponse)
	})

	Convey("When a filter blueprint does not exist, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().NotFound(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})

	Convey("When filter blueprint is unpublished, and the request is unauthenticated, a not found is returned", t, func() {
		r, err := http.NewRequest("GET", cfg().Host+"/filters/12345678", http.NoBody)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, filterNotFoundResponse)
	})
}

func TestSuccessfulUpdateFilterBlueprint_PublishedDataset(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Given a published dataset", t, func() {
		testETagUpdated := "testETagUpdated"
		w := httptest.NewRecorder()

		mockDatastore := &apimock.DataStoreMock{
			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
				return nil
			},
			GetFilterFunc: func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
				if eTagSelector != mongo.AnyETag && eTagSelector != testETag {
					return nil, filters.ErrFilterBlueprintConflict
				}
				return &models.Filter{Dataset: &models.Dataset{ID: "123", Edition: "2017", Version: 1}, InstanceID: "12345678", Published: &models.Published, Dimensions: []models.Dimension{{Name: "time", Options: []string{"2014", "2015"}}, {Name: "1_age"}}, ETag: testETag}, nil
			},
			UpdateFilterFunc: func(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
				if eTagSelector != testETag {
					return "", filters.ErrFilterBlueprintConflict
				}
				return testETagUpdated, nil
			},
		}

		filterAPI := api.Setup(cfg(), mux.NewRouter(), mockDatastore, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

		Convey("When a PUT request is made to the filters endpoint and a valid ETag", func() {
			reader := strings.NewReader(`{"dataset":{"version":1}}`)
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
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
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is not called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
			})

			Convey("Then the request body has been drained", func() {
				bytesRead, err := r.Body.Read(make([]byte, 1))
				So(bytesRead, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)
			})
		})

		Convey("When a PUT request is made to the filters endpoint to submit a filter blueprint", func() {
			reader := strings.NewReader("{}")
			r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312?submitted=true", reader)
			r.Header.Set("If-Match", testETag)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("Then the data store is called to create a new filter output", func() {
				So(len(mockDatastore.CreateFilterOutputCalls()), ShouldEqual, 1)

				filterOutput := mockDatastore.CreateFilterOutputCalls()[0]
				So(len(filterOutput.Filter.Events), ShouldEqual, 1)

				So(filterOutput.Filter.Events[0].Type, ShouldEqual, models.EventFilterOutputCreated)
			})

			Convey("Then the response is 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the updated ETag is returned in the ETag header", func() {
				So(w.Result().Header.Get("ETag"), ShouldResemble, testETagUpdated)
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
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("Successfully send a request to submit an unpublished filter blueprint", t, func() {
		reader := strings.NewReader("{}")
		r := createAuthenticatedRequest("PUT", cfg().Host+"/filters/21312?submitted=true", reader)
		r.Header.Set("If-Match", testETag)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().Unpublished(), &mock.FilterJob{}, mock.NewDatasetAPI().Unpublished(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Result().Header.Get("ETag"), ShouldResemble, testETag1)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})
}

func TestFailedToUpdateFilterBlueprint(t *testing.T) {
	filterFlexAPIMock := &apimock.FilterFlexAPIMock{}

	t.Parallel()

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
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

	Convey("When an empty json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{}")
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
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

	Convey("When a json message is sent to update filter blueprint that doesn't exist, a status of not found is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
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

	Convey("When no authentication is provided to update an unpublished filter, a not found is returned", t, func() {
		reader := strings.NewReader(`{"dimensions":[]}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
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

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the version does not exist, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, mock.NewDatasetAPI().VersionNotFound(), filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, versionNotFoundResponse)

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the datset version of a filter blueprint and the current dimensions do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimensions chosen: [time 1_age]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a json message is sent to change the dataset version of a filter blueprint and the current dimension options do not match, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":2}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		r.Header.Set("If-Match", testETag)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
		filterAPI.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Result().Header.Get("ETag"), ShouldResemble, "")

		response := w.Body.String()
		So(response, ShouldResemble, "incorrect dimension options chosen: [28]\n")

		Convey("Then the request body has been drained", func() {
			bytesRead, err := r.Body.Read(make([]byte, 1))
			So(bytesRead, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("When a request is made without providing an If-Match header, a status of bad request is returned", t, func() {
		reader := strings.NewReader(`{"dataset":{"version":1}}`)
		r, err := http.NewRequest("PUT", cfg().Host+"/filters/21312", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		filterAPI := api.Setup(cfg(), mux.NewRouter(), mock.NewDataStore().InvalidDimensionOption(), &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)
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

func TestRemoveDuplicatesAndEmptyOptions(t *testing.T) {
	Convey("Given a string array with duplicate options", t, func() {
		duplicates := []string{"1", "2", "2", "2", "abcde", "abd", "abcde"}

		Convey("When I call remove duplicate and empty function", func() {
			withoutDuplicates := api.RemoveDuplicateAndEmptyOptions(duplicates)

			Convey("Then the duplicates are removed", func() {
				expected := []string{"1", "2", "abcde", "abd"}
				So(withoutDuplicates, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a string array with empty options", t, func() {
		duplicates := []string{"", "1", "", "2", "", "3"}

		Convey("When I call remove duplicate and empty option function", func() {
			withoutEmpty := api.RemoveDuplicateAndEmptyOptions(duplicates)

			Convey("Then the empty values are removed", func() {
				expected := []string{"1", "2", "3"}
				So(withoutEmpty, ShouldResemble, expected)
			})
		})
	})
}

func TestRequestForwardingMiddleware(t *testing.T) {
	Convey("Given the assert dataset feature flag is toggled on", t, func() {
		conf := cfg()
		conf.AssertDatasetType = true

		w := httptest.NewRecorder()

		Convey("When a POST request is made to the filters endpoint and the dataset type is cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "cantabular_flexible_table",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, datasetAPIMock, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is made to check the dataset type", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters endpoint and the dataset type is not cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "other",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, datasetAPIMock, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "other"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is made to check the dataset type", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
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

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
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

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
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

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is not flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
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

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When there is a PUT request to /filter-outputs/test-output-id and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check the filter output type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
			})
			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})
		Convey("When there is a PUT request to /filter-outputs/test-output-id and the filter type is NOT flexible", func() {
			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "other"} }`)
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "NOT-flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			datastoreMock.CreateFilterOutputFunc = func(ctx context.Context, filter *models.Filter) error { return nil }

			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check the filter type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 1)
			})
			Convey("The request is NOT forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
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
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When there is a PUT request to /filters/{id} and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("PUT", cfg().Host+"/filters/test-output-id", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is made to datastore to check its filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
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
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
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
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is not flexible", func() {
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
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/{id}/dimensions/{dimension}/options/{option} endpoint and the filter type is flexible", func() {
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

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 1)
			})

			Convey("The request is forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a GET request is made to the filters/{id}/dimensions/{dimension}/options/{option} endpoint and the filter type is not flexible", func() {
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
					Type: "not-flexible",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is made to check the filter type", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 2)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given the assert dataset feature flag is toggled off", t, func() {
		conf := cfg()
		conf.AssertDatasetType = false

		w := httptest.NewRecorder()

		Convey("When a POST request is made to the filters endpoint and the dataset type is cantabular_flexible_table", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datasetAPIMock := mock.NewDatasetAPI().Mock

			datasetAPIMock.GetFunc = func(ctx context.Context, ut, st, cid, dsid string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					Type: "cantabular_flexible_table",
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("POST", cfg().Host+"/filters", reader)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to dataset-api is not made", func() {
				So(len(datasetAPIMock.GetCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions endpoint and the filter type is flexible", func() {
			filterFlexAPIMock := &apimock.FilterFlexAPIMock{
				ForwardRequestFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader([]byte("test body"))),
						StatusCode: 200,
					}, nil
				},
			}

			datastoreMock := mock.NewDataStore().Mock

			datastoreMock.GetFilterFunc = func(ctx context.Context, filterID, etag string) (*models.Filter, error) {
				return &models.Filter{
					Type: "flexible",
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest("GET", cfg().Host+"/filters/foo/dimensions", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PUT request is made to the filter-outputs/id and the filter type is flexible", func() {
			filterFlexMock, datastoreMock := mock.GenerateMocksForMiddleware(http.StatusOK, 1, "flexible")

			filterAPI := api.Setup(conf, mux.NewRouter(), datastoreMock, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			reader := strings.NewReader(`{"dataset":{"version":1, "edition":"1", "id":"1", "type": "cantabular_flexible_table"} }`)
			r, err := http.NewRequest("PUT", cfg().Host+"/filter-outputs/test-output-id", reader)
			So(err, ShouldBeNil)

			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call is NOT made to datastore to check the filter output type", func() {
				So(len(datastoreMock.GetFilterOutputCalls()), ShouldEqual, 0)
			})

			Convey("The request is NOT forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to the datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a POST request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodPost, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a PATCH request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodPatch, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a DELETE request is made to the filters/id/dimensions/name endpoint and the filter type is flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodDelete, cfg().Host+"/filters/foo/dimensions/bar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})

		Convey("When a GET request is made to the filters/id/dimensions/name/options/{option} endpoint and the filter type is flexible", func() {
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
					Dataset: &models.Dataset{
						Version: 1,
					},
				}, nil
			}

			filterAPI := api.Setup(conf, mux.NewRouter(), &mock.DataStore{}, &mock.FilterJob{}, &mock.DatasetAPI{}, filterFlexAPIMock, hostURL, datasetAPIURL, parsedDownloadServiceURL, false)

			r, err := http.NewRequest(http.MethodGet, cfg().Host+"/filters/foo/dimensions/bar/options/foobar", http.NoBody)
			So(err, ShouldBeNil)
			filterAPI.Router.ServeHTTP(w, r)

			Convey("A call to datastore is not made", func() {
				So(len(datastoreMock.GetFilterCalls()), ShouldEqual, 0)
			})

			Convey("The request is not forwarded to dp-cantabular-filter-flex-api", func() {
				So(len(filterFlexAPIMock.ForwardRequestCalls()), ShouldEqual, 0)
			})
		})
	})
}
