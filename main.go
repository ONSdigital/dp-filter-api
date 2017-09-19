package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filterJobQueue"
	"github.com/ONSdigital/dp-filter-api/postgres"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	_ "github.com/lib/pq"
)

func main() {
	log.Namespace = "dp-filter-api"

	signals := make(chan os.Signal)
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

	db, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		log.ErrorC("DB open error", err, nil)
		os.Exit(1)
	}

	dataStore, err := postgres.NewDatastore(db)
	if err != nil {
		log.ErrorC("Create postgres error", err, nil)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Brokers, cfg.FilterJobSubmittedTopic, int(envMax))
	if err != nil {
		log.ErrorC("Create kafka producer error", err, nil)
		os.Exit(1)
	}

	jobQueue := filterJobQueue.CreateJobQueue(producer.Output())

	apiErrors := make(chan error)

	api.CreateFilterAPI(cfg.SecretKey, cfg.Host, cfg.BindAddr, dataStore, &jobQueue, apiErrors)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() {
		log.Info(fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), nil)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		api.Close(ctx)
		// Close producer after http server has closed so if a message
		// needs to be sent to kafka off a request it can
		err := producer.Close(ctx)
		if err != nil {
			log.Error(err, nil)
		}

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
