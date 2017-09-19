package api

import (
	"context"
	"sync"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

const (
	gracefulShutdownMsg = "graceful shutdown of http server complete"
)

var httpServer *server.Server
var serverErrors chan error
var once sync.Once

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
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(secretKey, host, bindAddr string, datastore DataStore, jobQueue JobQueue, errorChan chan error) {
	router := mux.NewRouter()
	_ = routes(secretKey, host, router, datastore, jobQueue)

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
func routes(secretKey, host string, router *mux.Router, dataStore DataStore, jobQueue JobQueue) *FilterAPI {
	api := FilterAPI{internalToken: secretKey, host: host, dataStore: dataStore, router: router, jobQueue: jobQueue}

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

func Close(ctx context.Context) {
	httpServer.Shutdown(ctx)
	log.Info(gracefulShutdownMsg, nil)
}
