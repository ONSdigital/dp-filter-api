package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filterflex"
	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filteroutputqueue"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Service contains all the configs, server and clients to run the Dataset API
type Service struct {
	Cfg                           *config.Config
	FilterStore                   MongoDB
	FilterOutputSubmittedProducer kafka.IProducer
	IdentityClient                *identity.Client
	datasetAPI                    *dataset.Client
	filterFlexAPI                 *filterflex.Client
	HealthCheck                   HealthChecker
	Server                        HTTPServer
	api                           *api.FilterAPI
}

// GetFilterStore returns an initialised connection to filter store (mongo database)
var GetFilterStore = func(cfg *config.Config) (datastore MongoDB, err error) {
	return mongo.CreateFilterStore(cfg.MongoConfig, cfg.Host)
}

// GetProducer returns a kafka producer
var GetProducer = func(ctx context.Context, cfg *config.Config, kafkaBrokers []string, topic string) (kafkaProducer kafka.IProducer, err error) {
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
	}
	if cfg.KafkaSecProtocol == "TLS" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}
	producerChannels := kafka.CreateProducerChannels()
	return kafka.NewProducer(ctx, kafkaBrokers, topic, producerChannels, pConfig)
}

// GetHealthCheck returns a healthcheck
var GetHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) HealthChecker {
	hc := healthcheck.New(version, criticalTimeout, interval)
	return &hc
}

// GetHTTPServer returns an http server
var GetHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
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
	svc.Cfg = cfg

	// Get data store.
	svc.FilterStore, err = GetFilterStore(svc.Cfg)
	if err != nil {
		log.Error(ctx, "could not connect to mongodb", err)
		// We don't return 'err' here because we don't want to stop this service
		// due to a failure in connecting with mongoDB.
		// A failing healthcheck Checker will be created later in registerCheckers().
	}

	// Get kafka producer
	svc.FilterOutputSubmittedProducer, err = GetProducer(ctx, cfg, svc.Cfg.Brokers, svc.Cfg.FilterOutputSubmittedTopic)
	if err != nil {
		log.Error(ctx, "error creating kafka filter output submitted producer", err)
		return err
	}

	// Create Identity Client
	if svc.Cfg.EnablePrivateEndpoints {
		svc.IdentityClient = identity.New(svc.Cfg.ZebedeeURL)
	}

	// Create Dataset API client.
	svc.datasetAPI = dataset.NewAPIClient(svc.Cfg.DatasetAPIURL)
	svc.filterFlexAPI = filterflex.New(filterflex.Config{
		HostURL: svc.Cfg.FilterFlexAPIURL,
	})

	// Get HealthCheck and register checkers
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "error creating version info", err)
		return err
	}
	svc.HealthCheck = GetHealthCheck(versionInfo, svc.Cfg.HealthCheckCriticalTimeout, svc.Cfg.HealthCheckInterval)
	if err := svc.registerCheckers(ctx); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Get HTTP router and server with middleware
	r := mux.NewRouter()
	r.Use(otelmux.Middleware(cfg.OTServiceName))
	m := svc.createMiddleware(ctx, cfg)
	svc.Server = GetHTTPServer(svc.Cfg.BindAddr, m.Then(r))

	// Create API, with outputQueue
	outputQueue := filteroutputqueue.CreateOutputQueue(svc.FilterOutputSubmittedProducer.Channels().Output)
	svc.api = api.Setup(
		svc.Cfg,
		r,
		svc.FilterStore,
		&outputQueue,
		svc.datasetAPI,
		svc.filterFlexAPI,
	)
	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {
	// Start kafka logging
	svc.FilterOutputSubmittedProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+svc.Cfg.FilterOutputSubmittedTopic)

	// Start healthcheck
	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		log.Info(ctx, "Starting api...")
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
}

// CreateMiddleware creates an Alice middleware chain of handlers
func (svc *Service) createMiddleware(ctx context.Context, cfg *config.Config) alice.Chain {
	healthCheckHandler := newMiddleware(svc.HealthCheck.Handler, "/health")
	middlewareChain := alice.New(
		healthCheckHandler,
		dphandlers.CheckHeader(dphandlers.CollectionID),
		otelhttp.NewMiddleware(cfg.OTServiceName),
	)

	if svc.Cfg.EnablePrivateEndpoints {
		log.Info(ctx, "private endpoints are enabled. using identity middleware")
		identityHandler := dphandlers.IdentityWithHTTPClient(svc.IdentityClient)
		middlewareChain = middlewareChain.Append(identityHandler)
	}

	return middlewareChain
}

// newMiddleware creates a new http.Handler to intercept /health requests.
func newMiddleware(healthcheckHandler func(http.ResponseWriter, *http.Request), endpoint string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			//nolint:goconst // "GET" is acceptable here
			if req.Method == "GET" && req.URL.Path == endpoint {
				healthcheckHandler(w, req)
				return
			} else if req.Method == "GET" && req.URL.Path == "/healthcheck" {
				http.NotFound(w, req)
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Cfg.ShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.HealthCheck != nil {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests
		if svc.Server != nil {
			if err := svc.Server.Shutdown(ctx); err != nil {
				log.Error(ctx, "failed to shutdown http server", err)
				hasShutdownError = true
			}
		}

		// Close MongoDB (if it exists)
		if svc.FilterStore != nil {
			log.Info(ctx, "closing mongoDB filter data store")
			if err := svc.FilterStore.Close(ctx); err != nil {
				// if err := mongolib.Close(ctx, svc.filterStore.Session); err != nil {
				log.Error(ctx, "unable to close mongo filter data store", err)
				hasShutdownError = true
			}
		}

		// Close Kafka Producer (it if exists)
		if svc.FilterOutputSubmittedProducer != nil {
			log.Info(ctx, "closing filter output submitted producer")
			if err := svc.FilterOutputSubmittedProducer.Close(ctx); err != nil {
				log.Error(ctx, "unable to close filter output submitted producer", err)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
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
			if updateErr := state.Update(healthcheck.StatusCritical, err.Error(), 0); updateErr != nil {
				log.Error(ctx, "failed to update healthcheck state", updateErr)
				return updateErr
			}
			return err
		}

		// set / register the failing healthcheck
		handler := criticalHandler
		if dependency != nil {
			// we have a dependency so instead register its Checker
			handler = dependency.Checker
		}
		if err = svc.HealthCheck.AddCheck(name, handler); err != nil {
			log.Error(ctx, fmt.Sprintf("error creating %s health check", strings.ToLower(name)), err)
			hasErrors = true
		}
	}

	registerChecker("Dataset API", svc.datasetAPI)
	registerChecker("Kafka Producer", svc.FilterOutputSubmittedProducer)
	registerChecker("Mongo DB", svc.FilterStore)

	// zebedee is used only for identity checking if private endpoints are enabled
	if svc.Cfg.EnablePrivateEndpoints {
		registerChecker("Zebedee", svc.IdentityClient)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
