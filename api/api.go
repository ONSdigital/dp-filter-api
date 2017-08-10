package api

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gorilla/mux"
)

// JobQueue - An interface used to queue import jobs
type JobQueue interface {
	Queue(job *models.Filter) error
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host      string
	dataStore DataStore
	router    *mux.Router
	jobQueue  JobQueue
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(host string, router *mux.Router, dataStore DataStore, jobQueue JobQueue) *FilterAPI {
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthCheck)

	api := FilterAPI{host: host, dataStore: dataStore, router: router, jobQueue: jobQueue}
	api.router.HandleFunc("/filters", api.addFilterJob).Methods("POST")
	api.router.HandleFunc("/filters/{filter_job_id}", api.getFilterJob).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}", api.updateFilterJob).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions", api.getFilterJobDimensions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}", api.getFilterJobDimension).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}", api.addFilterJobDimension).Methods("POST")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options", api.getFilterJobDimensionOptions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options/{option}", api.getFilterJobDimensionOption).Methods("GET")
	api.router.HandleFunc("/filters/{filter_job_id}/dimensions/{name}/options/{option}", api.addFilterJobDimensionOption).Methods("POST")
	return &api
}
