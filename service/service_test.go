package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/service/mock"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "1599210455"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errMongo    = errors.New("MongoDB error")
	errGraph    = errors.New("GraphDB error")
	errKafka    = errors.New("Kafka producer error")
	errServer   = errors.New("HTTP Server error")
	errAddCheck = errors.New("healthcheck add check error")
)

func TestNew(t *testing.T) {
	Convey("New returns a new uninitialised service", t, func() {
		So(New(), ShouldResemble, &Service{})
	})
}

func TestInit(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		cfg.EnablePrivateEndpoints = true
		So(err, ShouldBeNil)

		mongoDBMock := &mock.MongoDBMock{}
		getFilterStore = func(cfg *config.Config) (datastore MongoDB, err error) {
			return mongoDBMock, nil
		}

		graphMock := &graph.DB{Driver: &mock.GraphDriverMock{}}
		graphErrorConsumerMock := &mock.CloserMock{CloseFunc: func(ctx context.Context) error {
			return nil
		}}
		getObservationStore = func(ctx context.Context) (observationStore *graph.DB, graphDBErrorConsumer Closer, err error) {
			return graphMock, graphErrorConsumerMock, nil
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}
		getProducer = func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafkaProducer kafka.IProducer, err error) {
			return kafkaProducerMock, nil
		}

		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		}
		getHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) HealthChecker {
			return hcMock
		}

		serverMock := &mock.HTTPServerMock{}
		getHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
			return serverMock
		}

		svc := &Service{}

		Convey("Given that initialising MongoDB datastore returns an error", func() {
			getFilterStore = func(cfg *config.Config) (datastore MongoDB, err error) {
				return nil, errMongo
			}

			Convey("Then service Init succeeds, MongoDB datastore dependency is not set and further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldBeNil)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldResemble, serverMock)

				Convey("And all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 5)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Graph DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that initialising GraphDB observation store returns an error", func() {
			getObservationStore = func(ctx context.Context) (observationStore *graph.DB, graphErrorConsumer Closer, err error) {
				return nil, nil, errGraph
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errGraph)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldBeNil)
				So(svc.filterOutputSubmittedProducer, ShouldBeNil)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("Given that initialising the kafka Producer returns an error", func() {
			getProducer = func(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafkaProducer kafka.IProducer, err error) {
				return nil, errKafka
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errKafka)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldBeNil)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("Given that healthcheck versionInfo cannot be created due to a wrong build time", func() {
			wrongBuildTime := "wrongFormat"

			Convey("Then service Init fails with the expected error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, wrongBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "failed to parse build time")
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.healthCheck, ShouldBeNil)
				So(svc.server, ShouldBeNil)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			hcMock.AddCheckFunc = func(name string, checker healthcheck.Checker) error { return errAddCheck }

			Convey("Then service Init fails with the expected error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "unable to register checkers: Error(s) registering checkers for healthcheck")
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldBeNil)

				Convey("But all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 5)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Graph DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			Convey("Then service Init succeeds, all dependencies are initialised", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldResemble, serverMock)

				Convey("And all checks are registered", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 5)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Graph DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised and private endpoints are disabled", func() {
			cfg.EnablePrivateEndpoints = false

			Convey("Then service Init succeeds, all dependencies are initialised except identity client", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.filterStore, ShouldResemble, mongoDBMock)
				So(svc.observationStore, ShouldResemble, graphMock)
				So(svc.filterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.healthCheck, ShouldResemble, hcMock)
				So(svc.server, ShouldResemble, serverMock)
				So(svc.identityClient, ShouldBeNil)

				Convey("And all checks are registered, except Zebedee", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 4)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Graph DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Mongo DB")
				})
			})
		})
	})
}

func TestStart(t *testing.T) {

	Convey("Having a correctly initialised Service with mocked dependencies", t, func() {

		cfg, err := config.Get()
		cfg.EnablePrivateEndpoints = true
		So(err, ShouldBeNil)

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}

		hcMock := &mock.HealthCheckerMock{
			StartFunc: func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &mock.HTTPServerMock{}

		svc := &Service{
			cfg:                           cfg,
			filterOutputSubmittedProducer: kafkaProducerMock,
			healthCheck:                   hcMock,
			server:                        serverMock,
		}

		Convey("When a service with a successful HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return nil
			}
			serverWg.Add(1)
			svc.Start(ctx, make(chan error, 1))

			Convey("Then healthcheck is started and HTTP server starts listening", func() {
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a service with a failing HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return errServer
			}
			errChan := make(chan error, 1)
			serverWg.Add(1)
			svc.Start(ctx, errChan)

			Convey("Then HTTP server errors are reported to the provided errors channel", func() {
				rxErr := <-errChan
				So(rxErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service with mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mock.HealthCheckerMock{
			StopFunc: func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server was stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		// funcClose succeeds only if healthcheck and server are already stopped
		funcClose := func(ctx context.Context) error {
			if !hcStopped {
				return errors.New("Dependency was closed before healthcheck")
			}
			if !serverStopped {
				return errors.New("Dependency was closed before http server")
			}
			return nil
		}

		// mongoDB will fail if healthcheck or http server are not stopped
		mongoMock := &mock.MongoDBMock{
			CloseFunc: funcClose,
		}

		// graphDB will fail if healthcheck or http server are not stopped
		graphDriverMock := &mock.GraphDriverMock{
			CloseFunc: funcClose,
		}

		graphErrorConsumerMock := &mock.CloserMock{
			CloseFunc: funcClose,
		}

		// Kafka producer will fail if healthcheck or http server are not stopped
		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
			CloseFunc: funcClose,
		}

		svc := &Service{
			cfg:                           cfg,
			healthCheck:                   hcMock,
			server:                        serverMock,
			filterStore:                   mongoMock,
			observationStore:              &graph.DB{Driver: graphDriverMock},
			graphDBErrorConsumer:          graphErrorConsumerMock,
			filterOutputSubmittedProducer: kafkaProducerMock,
		}

		Convey("Given that all dependencies succeed to close", func() {
			Convey("Closing the service results in all the initialised dependencies being closed in the expected order", func() {
				err = svc.Close(context.Background())
				So(err, ShouldBeNil)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
				So(len(graphDriverMock.CloseCalls()), ShouldEqual, 1)
				So(len(graphErrorConsumerMock.CloseCalls()), ShouldEqual, 1)
				So(len(kafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that all dependencies fail to close", func() {
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				return errServer
			}
			mongoMock.CloseFunc = func(ctx context.Context) error {
				return errMongo
			}
			graphDriverMock.CloseFunc = func(ctx context.Context) error {
				return errGraph
			}
			graphErrorConsumerMock.CloseFunc = func(ctx context.Context) error {
				return errGraph
			}
			kafkaProducerMock.CloseFunc = func(ctx context.Context) error {
				return errKafka
			}

			Convey("Then closing the service fails with the expected error and further dependencies are attempted to close", func() {
				err = svc.Close(context.Background())
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
				So(len(graphDriverMock.CloseCalls()), ShouldEqual, 1)
				So(len(graphErrorConsumerMock.CloseCalls()), ShouldEqual, 1)
				So(len(kafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that a dependency takes more time to close than the graceful shutdown timeout", func() {
			cfg.ShutdownTimeout = 1 * time.Millisecond
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(2 * time.Millisecond)
				return nil
			}

			Convey("Then closing the service fails with context.DeadlineExceeded error and no further dependencies are attempted to close", func() {
				err = svc.Close(context.Background())
				So(err, ShouldResemble, context.DeadlineExceeded)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 0)
				So(len(graphDriverMock.CloseCalls()), ShouldEqual, 0)
				So(len(graphErrorConsumerMock.CloseCalls()), ShouldEqual, 0)
				So(len(kafkaProducerMock.CloseCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Having a non-initialised service", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		svc := &Service{
			cfg: cfg,
		}

		Convey("Closing the service succeeds without attempting to close any non-initialised dependency", func() {
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
		})
	})
}
