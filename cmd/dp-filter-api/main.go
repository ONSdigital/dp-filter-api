package main

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-filter-api/kafkaadapter"
	mongolib "github.com/ONSdigital/dp-mongodb"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filterOutputQueue"
	"github.com/ONSdigital/dp-filter-api/initialise"
	"github.com/ONSdigital/dp-filter-api/preview"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/log.go/log"
)

func main() {

	log.Namespace = "dp-filter-api"
	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(ctx, err, "unable to retrieve configuration")

	// sensitive fields are omitted from config.String()
	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": cfg})

	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	exitIfError(ctx, err, "encountered error parsing kafka max bytes")

	var serviceList initialise.ExternalServiceList

	dataStore, err := serviceList.GetFilterStore(cfg)
	logIfError(ctx, err, "could not connect to mongodb")

	observationStore, err := serviceList.GetObservationStore()
	logIfError(ctx, err, "could not connect to graph")

	producer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.FilterOutputSubmittedTopic,
		initialise.FilterOutputSubmittedProducer,
		int(envMax),
	)
	logIfError(ctx, err, "error creating kafka filter output submitted producer")
	producer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.FilterOutputSubmittedTopic)

	var auditor audit.AuditorService
	var auditProducer *kafka.Producer

	if cfg.EnablePrivateEndpoints {
		log.Event(ctx, "private endpoints enabled, enabling auditing")

		auditProducer, err = serviceList.GetProducer(
			ctx,
			cfg.Brokers,
			cfg.AuditEventsTopic,
			initialise.AuditProducer,
			0,
		)
		logIfError(ctx, err, "error creating kafka audit producer")
		auditProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.AuditEventsTopic)

		auditProducerAdapter := kafkaadapter.NewProducerAdapter(auditProducer)
		auditor = audit.New(auditProducerAdapter, "dp-filter-api")
	} else {
		log.Event(ctx, "private endpoints disabled, auditing will not be enabled")
		auditor = &audit.NopAuditor{}
	}

	// todo: remove config.DatasetAPIAuthToken when the DatasetAPI supports identity based auth.
	datasetAPI := dataset.NewAPIClient(cfg.DatasetAPIURL)

	previewDatasets := preview.DatasetStore{Store: observationStore}
	outputQueue := filterOutputQueue.CreateOutputQueue(producer.Channels().Output)

	//healthTicker := healthcheck.NewTicker(
	//	cfg.HealthCheckInterval,
	//	observationStore,
	//	mongolib.NewHealthCheckClient(dataStore.Session),
	//	datasetAPI,
	//)

	apiErrors := make(chan error, 1)

	api.CreateFilterAPI(ctx, cfg.Host,
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

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Event(ctx, "os signal received")
	}

	log.Event(ctx, fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), log.INFO)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		if err = api.Close(ctx); err != nil {
			logIfError(ctx, err, "unable to close api server")
		}

		//healthTicker.Close()

		if serviceList.FilterStore {
			log.Event(ctx, "closing filter store")
			// mongo.Close() may use all remaining time in the context
			logIfError(ctx, mongolib.Close(ctx, dataStore.Session), "unable to close filter store")
		}

		if serviceList.ObservationStore {
			log.Event(ctx, "closing observation store")
			logIfError(ctx, observationStore.Close(ctx), "unable to close observation store")
		}

		if serviceList.FilterOutputSubmittedProducer {
			log.Event(ctx, "closing filter output submitted producer")
			// Close producer after http server has closed so if a message
			// needs to be sent to kafka off a request it can
			logIfError(ctx, producer.Close(ctx), "unable to close filter output submitted producer")
		}

		if serviceList.AuditProducer {
			log.Event(ctx, "closing audit producer")
			logIfError(ctx, auditProducer.Close(ctx), "unable to close audit producer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "Shutdown complete")
	os.Exit(1)
}

func exitIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Event(ctx, message, log.Error(err))
		os.Exit(1)
	}
}

func logIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Event(ctx, message, log.Error(err))
	}
}
