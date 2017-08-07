package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filterJobQueue"
	"github.com/ONSdigital/dp-filter-api/postgres"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	log.Namespace = "dp-filter-api"

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
	router := mux.NewRouter()

	_ = api.CreateFilterAPI(cfg.Host, router, dataStore, &jobQueue)
	if err = http.ListenAndServe(cfg.BindAddr, router); err != nil {
		log.Error(err, log.Data{"BIND_ADDR": cfg.BindAddr})
	}

	producer.Closer() <- true
}
