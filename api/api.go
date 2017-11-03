package api

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// OutputQueue - An interface used to queue filter outputs
type OutputQueue interface {
	Queue(output *models.Filter) error
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host          string
	dataStore     DataStore
	internalToken string
	outputQueue   OutputQueue
	router        *mux.Router
	datasetAPI    DatasetAPIer
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(secretKey, host, bindAddr string, datastore DataStore, outputQueue OutputQueue, errorChan chan error, datasetAPI DatasetAPIer) {
	router := mux.NewRouter()
	routes(secretKey, host, router, datastore, outputQueue, datasetAPI)

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
func routes(secretKey, host string, router *mux.Router, dataStore DataStore, outputQueue OutputQueue, datasetAPI DatasetAPIer) *FilterAPI {
	api := FilterAPI{internalToken: secretKey, host: host, dataStore: dataStore, router: router, outputQueue: outputQueue, datasetAPI: datasetAPI}

	router.Path("/healthcheck").Methods("GET").HandlerFunc(api.healthCheck)

	api.router.HandleFunc("/filters", api.addFilterBlueprint).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.getFilterBlueprint).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.updateFilterBlueprint).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions", api.getFilterBlueprintDimensions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.getFilterBlueprintDimension).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.addFilterBlueprintDimension).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.removeFilterBlueprintDimension).Methods("DELETE")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options", api.getFilterBlueprintDimensionOptions).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.getFilterBlueprintDimensionOption).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.addFilterBlueprintDimensionOption).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.removeFilterBlueprintDimensionOption).Methods("DELETE")

	api.router.HandleFunc("/filter-outputs/{filter_output_id}", api.getFilterOutput).Methods("GET")
	api.router.HandleFunc("/filter-outputs/{filter_output_id}", api.updateFilterOutput).Methods("PUT")
	return &api
}

func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
