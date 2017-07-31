package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/postgres"
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

	router := mux.NewRouter()

	_ = api.CreateFilterAPI(cfg.Host, router, dataStore)
	err = http.ListenAndServe(cfg.BindAddr, router)
	if err != nil {
		log.Error(err, log.Data{"BIND_ADDR": cfg.BindAddr})
	}
}
