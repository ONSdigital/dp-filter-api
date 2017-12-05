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
	"github.com/ONSdigital/dp-filter-api/dataset"
	"github.com/ONSdigital/dp-filter-api/filterOutputQueue"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	mongoclosure "github.com/ONSdigital/go-ns/mongo"
	"github.com/ONSdigital/go-ns/rchttp"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func main() {
	log.Namespace = "filter-api"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	if err != nil {
		log.ErrorC("encountered error parsing kafka max bytes", err, nil)
		os.Exit(1)
	}

	dataStore, err := mongo.CreateFilterStore(cfg.MongoDBURL, cfg.Host)
	if err != nil {
		log.ErrorC("could not connect to mongodb", err, nil)
		os.Exit(1)
	}

	pool, err := bolt.NewClosableDriverPool(cfg.Neo4jURL, cfg.Neo4jPoolSize)
	if err != nil {
		log.ErrorC("could not connect to neo4j", err, nil)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Brokers, cfg.FilterOutputSubmittedTopic, int(envMax))
	if err != nil {
		log.ErrorC("Create kafka producer error", err, nil)
		os.Exit(1)
	}

	client := rchttp.DefaultClient
	datasetAPI := dataset.NewDatasetAPI(client, cfg.DatasetAPIURL, cfg.DatasetAPIAuthToken)

	observationStore := observation.NewStore(pool)
	previewDatasets := preview.PreviewDatasetStore{Store: observationStore}
	outputQueue := filterOutputQueue.CreateOutputQueue(producer.Output())

	apiErrors := make(chan error, 1)

	api.CreateFilterAPI(cfg.SecretKey, cfg.Host, cfg.BindAddr, dataStore, &outputQueue, apiErrors, datasetAPI, &previewDatasets)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() {
		log.Info(fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), nil)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

		if err := api.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		// mongo.Close() may use all remaining time in the context
		if err = mongoclosure.Close(ctx, dataStore.Session); err != nil {
			log.Error(err, nil)
		}

		if err = pool.Close(); err != nil {
			log.Error(err, nil)
		}

		// Close producer after http server has closed so if a message
		// needs to be sent to kafka off a request it can
		if err := producer.Close(ctx); err != nil {
			log.Error(err, nil)
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
