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
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/dp-graph/graph"
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
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Info("loaded config", log.Data{
		"config": cfg,
	})

	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	if err != nil {
		log.ErrorC("encountered error parsing kafka max bytes", err, nil)
		os.Exit(1)
	}

	dataStore, err := mongo.CreateFilterStore(cfg.MongoConfig, cfg.Host)
	if err != nil {
		log.ErrorC("could not connect to mongodb", err, nil)
		os.Exit(1)
	}

	observationStore, err := graph.New(context.Background(), graph.Subsets{Observation: true})
	if err != nil {
		log.ErrorC("could not connect to graph", err, nil)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Brokers, cfg.FilterOutputSubmittedTopic, int(envMax))
	if err != nil {
		log.ErrorC("Create kafka producer error", err, nil)
		os.Exit(1)
	}

	var auditor audit.AuditorService
	var auditProducer kafka.Producer

	if cfg.EnablePrivateEndpoints {
		log.Info("private endpoints enabled, enabling auditing", nil)

		auditProducer, err = kafka.NewProducer(cfg.Brokers, cfg.AuditEventsTopic, 0)
		if err != nil {
			log.ErrorC("error creating kafka audit producer", err, nil)
			os.Exit(1)
		}

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
		//	observationStore.Healthcheck(),
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

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() {
		log.Info(fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), nil)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

		if err = api.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		healthTicker.Close()

		// mongo.Close() may use all remaining time in the context
		if err = mongolib.Close(ctx, dataStore.Session); err != nil {
			log.Error(err, nil)
		}

		if err = observationStore.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		// Close producer after http server has closed so if a message
		// needs to be sent to kafka off a request it can
		if err := producer.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		if cfg.EnablePrivateEndpoints {
			log.Debug("exiting audit producer", nil)
			if err = auditProducer.Close(ctx); err != nil {
				log.Error(err, nil)
			}
		}

		cancel()
		log.Info("Shutdown complete", nil)
		os.Exit(1)
	}

	for {
		select {
		case err := <-producer.Errors():
			log.ErrorC("kafka producer error received", err, nil)
			gracefulShutdown()
		case err := <-apiErrors:
			log.ErrorC("api error received", err, nil)
			gracefulShutdown()
		case <-signals:
			log.Debug("os signal received", nil)
			gracefulShutdown()
		}
	}
}
