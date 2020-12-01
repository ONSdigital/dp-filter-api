package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filterOutputQueue"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Dataset API
type Service struct {
	cfg                           *config.Config
	filterStore                   MongoDB
	observationStore              *graph.DB
	graphDBErrorConsumer          Closer
	filterOutputSubmittedProducer kafka.IProducer
	identityClient                *identity.Client
	datasetAPI                    *dataset.Client
	healthCheck                   HealthChecker
	server                        HTTPServer
	api                           *api.FilterAPI
}

// getFilterStore returns an initialised connection to filter store (mongo database)
var getFilterStore = func(cfg *config.Config) (datastore MongoDB, err error) {
	return mongo.CreateFilterStore(cfg.MongoConfig, cfg.Host)
}

// getObservationStore returns an initialised connection to observation store (graph database)
var getObservationStore = func(ctx context.Context) (observationStore *graph.DB, errorConsumer Closer, err error) {
	observationStore, err = graph.New(context.Background(), graph.Subsets{Observation: true})
	if err != nil {
		return nil, nil, err
	}

	errorConsumer = graph.NewLoggingErrorConsumer(ctx, observationStore.ErrorChan())

	return observationStore, errorConsumer, nil
}

// getProducer returns a kafka producer
var getProducer = func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafkaProducer kafka.IProducer, err error) {
	producerChannels := kafka.CreateProducerChannels()
	return kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, producerChannels)
}

// getHealthCheck returns a healthcheck
var getHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) HealthChecker {
	hc := healthcheck.New(version, criticalTimeout, interval)
	return &hc
}

// getHTTPServer returns an http server
var getHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// New creates a new empty service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware
func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) (err error) {

	svc.cfg = cfg

	// Get data store.
	svc.filterStore, err = getFilterStore(svc.cfg)
	if err != nil {
		log.Event(ctx, "could not connect to mongodb", log.ERROR, log.Error(err))
		// We don't return 'err' here because we don't want to stop this service
		// due to a failure in connecting with mongoDB.
		// A failing healthcheck Checker will be created later in registerCheckers().
	}

	// Get observation store
	svc.observationStore, svc.graphDBErrorConsumer, err = getObservationStore(ctx)
	if err != nil {
		log.Event(ctx, "could not connect to graph", log.ERROR, log.Error(err))
		return err
	}

	// Get kafka producer
	svc.filterOutputSubmittedProducer, err = getProducer(ctx, svc.cfg.Brokers, svc.cfg.FilterOutputSubmittedTopic, svc.cfg.KafkaMaxBytes)
	if err != nil {
		log.Event(ctx, "error creating kafka filter output submitted producer", log.ERROR, log.Error(err))
		return err
	}

	// Create Identity Client
	if svc.cfg.EnablePrivateEndpoints {
		svc.identityClient = identity.New(svc.cfg.ZebedeeURL)
	}

	// Create Dataset API client.
	svc.datasetAPI = dataset.NewAPIClient(svc.cfg.DatasetAPIURL)

	// Get HealthCheck and register checkers
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "error creating version info", log.FATAL, log.Error(err))
		return err
	}
	svc.healthCheck = getHealthCheck(versionInfo, svc.cfg.HealthCheckCriticalTimeout, svc.cfg.HealthCheckInterval)
	if err := svc.registerCheckers(ctx); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Get HTTP router and server with middleware
	r := mux.NewRouter()
	m := svc.createMiddleware(ctx)
	svc.server = getHTTPServer(svc.cfg.BindAddr, m.Then(r))

	// Create API, with previewDatasets and outputQueue
	previewDatasets := preview.DatasetStore{Store: svc.observationStore}
	outputQueue := filterOutputQueue.CreateOutputQueue(svc.filterOutputSubmittedProducer.Channels().Output)
	svc.api = api.Setup(
		svc.cfg,
		r,
		svc.filterStore,
		&outputQueue,
		svc.datasetAPI,
		&previewDatasets)
	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {

	// Start kafka logging
	svc.filterOutputSubmittedProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+svc.cfg.FilterOutputSubmittedTopic)

	// Start healthcheck
	svc.healthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := svc.server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
}

// CreateMiddleware creates an Alice middleware chain of handlers
func (svc *Service) createMiddleware(ctx context.Context) alice.Chain {
	healthCheckHandler := newMiddleware(svc.healthCheck.Handler, "/health")
	oldHealthCheckHandler := newMiddleware(svc.healthCheck.Handler, "/healthcheck")
	middlewareChain := alice.New(
		healthCheckHandler,
		oldHealthCheckHandler,
		dphandlers.CheckHeader(dphandlers.CollectionID))

	if svc.cfg.EnablePrivateEndpoints {
		log.Event(ctx, "private endpoints are enabled. using identity middleware", log.INFO)
		identityHandler := dphandlers.IdentityWithHTTPClient(svc.identityClient)
		middlewareChain = middlewareChain.Append(identityHandler)
	}

	return middlewareChain
}

// newMiddleware creates a new http.Handler to intercept /health requests.
func newMiddleware(healthcheckHandler func(http.ResponseWriter, *http.Request), endpoint string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" && req.URL.Path == endpoint {
				healthcheckHandler(w, req)
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.cfg.ShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.healthCheck != nil {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests
		if svc.server != nil {
			if err := svc.server.Shutdown(ctx); err != nil {
				log.Event(ctx, "failed to shutdown http server", log.ERROR)
				hasShutdownError = true
			}
		}

		// Close MongoDB (if it exists)
		if svc.filterStore != nil {
			log.Event(ctx, "closing mongoDB filter data store", log.INFO)
			if err := svc.filterStore.Close(ctx); err != nil {
				// if err := mongolib.Close(ctx, svc.filterStore.Session); err != nil {
				log.Event(ctx, "unable to close mongo filter data store", log.ERROR)
				hasShutdownError = true
			}
		}

		// Close GraphDB (if it exists)
		if svc.observationStore != nil {
			log.Event(ctx, "closing graph DB observation store", log.INFO)
			if err := svc.observationStore.Close(ctx); err != nil {
				log.Event(ctx, "unable to close graph DB observation store", log.ERROR)
				hasShutdownError = true
			}

			log.Event(ctx, "closing graph DB error consumer", log.INFO)
			if err := svc.graphDBErrorConsumer.Close(ctx); err != nil {
				log.Event(ctx, "unable to close graph DB error consumer", log.ERROR)
				hasShutdownError = true
			}
		}

		// Close Kafka Producer (it if exists)
		if svc.filterOutputSubmittedProducer != nil {
			log.Event(ctx, "closing filter output submitted producer", log.INFO)
			if err := svc.filterOutputSubmittedProducer.Close(ctx); err != nil {
				log.Event(ctx, "unable to close filter output submitted producer", log.ERROR)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the checkers for the service clients to the health check object.
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	// generic interface that must be satisfied by all health-checkable dependencies
	type Dependency interface {
		Checker(context.Context, *healthcheck.CheckState) error
	}

	// generic register checker method - if dependency is nil, a failing healthcheck will be created.
	registerChecker := func(name string, dependency Dependency) {
		criticalHandler := func(ctx context.Context, state *healthcheck.CheckState) error {
			err := errors.New(fmt.Sprintf("%s not initialised", strings.ToLower(name)))
			state.Update(healthcheck.StatusCritical, err.Error(), 0)
			return err
		}

		// set / register the failing healthcheck
		handler := criticalHandler
		if dependency != nil {
			// we have a dependency so instead register its Checker
			handler = dependency.Checker
		}
		if err = svc.healthCheck.AddCheck(name, handler); err != nil {
			log.Event(ctx, fmt.Sprintf("error creating %s health check", strings.ToLower(name)), log.ERROR, log.Error(err))
			hasErrors = true
		}
	}

	registerChecker("Dataset API", svc.datasetAPI)
	registerChecker("Kafka Producer", svc.filterOutputSubmittedProducer)
	registerChecker("Graph DB", svc.observationStore)
	registerChecker("Mongo DB", svc.filterStore)

	// zebedee is used only for identity checking if private endpoints are enabled
	if svc.cfg.EnablePrivateEndpoints {
		registerChecker("Zebedee", svc.identityClient)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
