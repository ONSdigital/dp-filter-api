package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filterOutputQueue"
	"github.com/ONSdigital/dp-filter-api/initialise"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	mongolib "github.com/ONSdigital/go-ns/mongo"
)

func main() {

	log.Namespace = "dp-filter-api"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(err, "unable to retrieve configuration")

	// sensitive fields are omitted from config.String()
	log.Info("loaded config", log.Data{
		"config": cfg,
	})

	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	exitIfError(err, "encountered error parsing kafka max bytes")

	var serviceList initialise.ExternalServiceList

	dataStore, err := serviceList.GetFilterStore(cfg)
	logIfError(err, "could not connect to mongodb")

	observationStore, err := serviceList.GetObservationStore()
	logIfError(err, "could not connect to graph")

	producer, err := serviceList.GetProducer(
		cfg.Brokers,
		cfg.FilterOutputSubmittedTopic,
		initialise.FilterOutputSubmitted,
		int(envMax),
	)
	logIfError(err, "error creating kafka filter output submitted producer")

	var auditor audit.AuditorService
	var auditProducer kafka.Producer

	if cfg.EnablePrivateEndpoints {
		log.Info("private endpoints enabled, enabling auditing", nil)

		auditProducer, err = serviceList.GetProducer(
			cfg.Brokers,
			cfg.AuditEventsTopic,
			initialise.Audit,
			0,
		)
		logIfError(err, "error creating kafka audit producer")

		auditor = audit.New(auditProducer, "dp-filter-api")
	} else {
		log.Info("private endpoints disabled, auditing will not be enabled", nil)
		auditor = &audit.NopAuditor{}
	}

	// todo: remove config.DatasetAPIAuthToken when the DatasetAPI supports identity based auth.
	datasetAPI := dataset.NewAPIClient(cfg.DatasetAPIURL, cfg.ServiceAuthToken, "")

	previewDatasets := preview.DatasetStore{Store: observationStore}
	outputQueue := filterOutputQueue.CreateOutputQueue(producer.Output())

	healthTicker := healthcheck.NewTicker(
		cfg.HealthCheckInterval,
		observationStore,
		mongolib.NewHealthCheckClient(dataStore.Session),
		datasetAPI,
	)

	apiErrors := make(chan error, 1)

	api.CreateFilterAPI(cfg.Host,
		cfg.BindAddr,
		cfg.ZebedeeURL,
		dataStore,
		&outputQueue,
		apiErrors,
		datasetAPI,
		&previewDatasets,
		cfg.EnablePrivateEndpoints,
		cfg.DownloadServiceURL,
		cfg.DownloadServiceSecretKey,
		auditor,
	)

	go func() {
		var producerErrors, auditProducerError chan (error)

		if serviceList.FilterOutputSubmittedProducer {
			producerErrors = producer.Errors()
		} else {
			producerErrors = make(chan error, 1)
		}

		if serviceList.AuditProducer {
			auditProducerError = auditProducer.Errors()
		} else {
			auditProducerError = make(chan error, 1)
		}

		select {
		case err := <-producerErrors:
			log.ErrorC("kafka producer error received", err, nil)
		case err := <-auditProducerError:
			log.ErrorC("kafka audit producer error received", err, nil)
		case err := <-apiErrors:
			log.ErrorC("api error received", err, nil)
		}
	}()

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Info("os signal received", nil)
	}

	log.Info(fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), nil)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// Close ticker first as it depends on other services/clients being available
		// Helps to prevent race conditions between health ticker/checker and graceful shutdown
		healthTicker.Close()

		if err = api.Close(ctx); err != nil {
			logIfError(err, "unable to close api server")
		}

		if serviceList.FilterStore {
			log.Info("closing filter store", nil)
			// mongo.Close() may use all remaining time in the context
			logIfError(mongolib.Close(ctx, dataStore.Session), "unable to close filter store")
		}

		if serviceList.ObservationStore {
			log.Info("closing observation store", nil)
			logIfError(observationStore.Close(ctx), "unable to close observation store")
		}

		if serviceList.FilterOutputSubmittedProducer {
			log.Info("closing filter output submitted producer", nil)
			// Close producer after http server has closed so if a message
			// needs to be sent to kafka off a request it can
			logIfError(producer.Close(ctx), "unable to close filter output submitted producer")
		}

		if serviceList.AuditProducer {
			log.Info("closing audit producer", nil)
			logIfError(auditProducer.Close(ctx), "unable to close audit producer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Info("Shutdown complete", nil)
	os.Exit(1)
}

func exitIfError(err error, message string) {
	if err != nil {
		log.ErrorC(message, err, nil)
		os.Exit(1)
	}
}

func logIfError(err error, message string) {
	if err != nil {
		log.ErrorC(message, err, nil)
	}
}
