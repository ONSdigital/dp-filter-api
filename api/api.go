package api

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// JobQueue - An interface used to queue import jobs
type JobQueue interface {
	Queue(job *models.Filter) error
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host          string
	dataStore     DataStore
	internalToken string
	jobQueue      JobQueue
	router        *mux.Router
	datasetAPI    DatasetAPIer
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(secretKey, host, bindAddr string, datastore DataStore, jobQueue JobQueue, errorChan chan error, datasetAPI DatasetAPIer) {
	router := mux.NewRouter()
	routes(secretKey, host, router, datastore, jobQueue, datasetAPI)

	httpServer = server.New(bindAddr, router)
	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting api...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("api http server returned error", err, nil)
			errorChan <- err
		}
	}()
}

// routes contain all endpoints for API
func routes(secretKey, host string, router *mux.Router, dataStore DataStore, jobQueue JobQueue, datasetAPI DatasetAPIer) *FilterAPI {
	api := FilterAPI{internalToken: secretKey, host: host, dataStore: dataStore, router: router, jobQueue: jobQueue, datasetAPI: datasetAPI}

	router.Path("/healthcheck").Methods("GET").HandlerFunc(api.healthCheck)

	api.router.HandleFunc("/filters", api.addFilterJob).Methods("POST")
	api.router.HandleFunc("/filters/{filter_job_id}", api.getFilterJob).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}", api.updateFilterJob).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions", api.getFilterJobDimensions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}", api.getFilterJobDimension).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}", api.addFilterJobDimension).Methods("POST")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}", api.removeFilterJobDimension).Methods("DELETE")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options", api.getFilterJobDimensionOptions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options/{option}", api.getFilterJobDimensionOption).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options/{option}", api.addFilterJobDimensionOption).Methods("POST")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options/{option}", api.removeFilterJobDimensionOption).Methods("DELETE")
	return &api
}

func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
