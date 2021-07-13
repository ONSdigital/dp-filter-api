package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/service"
	serviceMock "github.com/ONSdigital/dp-filter-api/service/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
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
		So(service.New(), ShouldResemble, &service.Service{})
	})
}

func TestInit(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		cfg.EnablePrivateEndpoints = true
		So(err, ShouldBeNil)

		mongoDBMock := &serviceMock.MongoDBMock{}
		service.GetFilterStore = func(cfg *config.Config) (datastore service.MongoDB, err error) {
			return mongoDBMock, nil
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
		}
		service.GetProducer = func(ctx context.Context, cfg *config.Config, kafkaBrokers []string, topic string) (kafkaProducer kafka.IProducer, err error) {
			return kafkaProducerMock, nil
		}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		}
		service.GetHealthCheck = func(version healthcheck.VersionInfo, criticalTimeout, interval time.Duration) service.HealthChecker {
			return hcMock
		}

		serverMock := &serviceMock.HTTPServerMock{}
		service.GetHTTPServer = func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		svc := &service.Service{}

		Convey("Given that initialising MongoDB datastore returns an error", func() {
			service.GetFilterStore = func(cfg *config.Config) (datastore service.MongoDB, err error) {
				return nil, errMongo
			}

			Convey("Then service Init succeeds, MongoDB datastore dependency is not set and further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldBeNil)
				So(svc.FilterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.Server, ShouldResemble, serverMock)

				Convey("And all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 4)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that initialising the kafka Producer returns an error", func() {
			service.GetProducer = func(ctx context.Context, cfg *config.Config, kafkaBrokers []string, topic string) (kafkaProducer kafka.IProducer, err error) {
				return nil, errKafka
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldResemble, errKafka)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldResemble, mongoDBMock)
				So(svc.FilterOutputSubmittedProducer, ShouldBeNil)
				So(svc.HealthCheck, ShouldBeNil)
				So(svc.Server, ShouldBeNil)
			})
		})

		Convey("Given that healthcheck versionInfo cannot be created due to a wrong build time", func() {
			wrongBuildTime := "wrongFormat"

			Convey("Then service Init fails with the expected error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, wrongBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "failed to parse build time")
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldResemble, mongoDBMock)
				So(svc.FilterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.HealthCheck, ShouldBeNil)
				So(svc.Server, ShouldBeNil)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			hcMock.AddCheckFunc = func(name string, checker healthcheck.Checker) error { return errAddCheck }

			Convey("Then service Init fails with the expected error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "unable to register checkers: Error(s) registering checkers for healthcheck")
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldResemble, mongoDBMock)
				So(svc.FilterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.Server, ShouldBeNil)

				Convey("But all checks try to register", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 4)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			Convey("Then service Init succeeds, all dependencies are initialised", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldResemble, mongoDBMock)
				So(svc.FilterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.Server, ShouldResemble, serverMock)

				Convey("And all checks are registered", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 4)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Mongo DB")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Zebedee")
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised and private endpoints are disabled", func() {
			cfg.EnablePrivateEndpoints = false

			Convey("Then service Init succeeds, all dependencies are initialised except identity client", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.FilterStore, ShouldResemble, mongoDBMock)
				So(svc.FilterOutputSubmittedProducer, ShouldResemble, kafkaProducerMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.Server, ShouldResemble, serverMock)
				So(svc.IdentityClient, ShouldBeNil)

				Convey("And all checks are registered, except Zebedee", func() {
					So(len(hcMock.AddCheckCalls()), ShouldEqual, 3)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Dataset API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka Producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Mongo DB")
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

		hcMock := &serviceMock.HealthCheckerMock{
			StartFunc: func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{}

		svc := &service.Service{
			Cfg:                           cfg,
			FilterOutputSubmittedProducer: kafkaProducerMock,
			HealthCheck:                   hcMock,
			Server:                        serverMock,
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
		hcMock := &serviceMock.HealthCheckerMock{
			StopFunc: func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
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
		mongoMock := &serviceMock.MongoDBMock{
			CloseFunc: funcClose,
		}

		// Kafka producer will fail if healthcheck or http server are not stopped
		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return &kafka.ProducerChannels{}
			},
			CloseFunc: funcClose,
		}

		svc := &service.Service{
			Cfg:                           cfg,
			HealthCheck:                   hcMock,
			Server:                        serverMock,
			FilterStore:                   mongoMock,
			FilterOutputSubmittedProducer: kafkaProducerMock,
		}

		Convey("Given that all dependencies succeed to close", func() {
			Convey("Closing the service results in all the initialised dependencies being closed in the expected order", func() {
				err = svc.Close(context.Background())
				So(err, ShouldBeNil)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
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
				So(len(kafkaProducerMock.CloseCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that a dependency takes more time to close than the graceful shutdown timeout", func() {
			cfg.ShutdownTimeout = 5 * time.Millisecond
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

			Convey("Then closing the service fails with context.DeadlineExceeded error and no further dependencies are attempted to close", func() {
				err = svc.Close(context.Background())
				So(err, ShouldResemble, context.DeadlineExceeded)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
				So(len(mongoMock.CloseCalls()), ShouldEqual, 0)
				So(len(kafkaProducerMock.CloseCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Having a non-initialised service", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		svc := &service.Service{
			Cfg: cfg,
		}

		Convey("Closing the service succeeds without attempting to close any non-initialised dependency", func() {
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
		})
	})
}
