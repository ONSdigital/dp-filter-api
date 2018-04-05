package api

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var httpServer *server.Server

type key string

const internalTokenKey key = "Internal-Token"

// OutputQueue - An interface used to queue filter outputs
type OutputQueue interface {
	Queue(output *models.Filter) error
}

//go:generate moq -out datastoretest/preview.go -pkg datastoretest . PreviewDataset

// PreviewDataset An interface used to generate previews
type PreviewDataset interface {
	GetPreview(filter *models.Filter, limit int) (*preview.FilterPreview, error)
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host        string
	dataStore   DataStore
	outputQueue OutputQueue
	router      *mux.Router
	datasetAPI  DatasetAPIer
	preview     PreviewDataset
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(host, bindAddr, zebedeeURL string,
	datastore DataStore,
	outputQueue OutputQueue,
	errorChan chan error,
	datasetAPI DatasetAPIer,
	preview PreviewDataset,
	enablePrivateEndpoints bool) {

	router := mux.NewRouter()
	routes(host, router, datastore, outputQueue, datasetAPI, preview, enablePrivateEndpoints)

	// Only add the identity middleware when running in publishing.
	if enablePrivateEndpoints {
		identityHandler := identity.Handler(true, zebedeeURL)
		alice := alice.New(identityHandler).Then(router)
		httpServer = server.New(bindAddr, alice)
	} else {
		httpServer = server.New(bindAddr, router)
	}

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
func routes(host string,
	router *mux.Router,
	dataStore DataStore,
	outputQueue OutputQueue,
	datasetAPI DatasetAPIer,
	preview PreviewDataset,
	enablePrivateEndpoints bool) *FilterAPI {

	api := FilterAPI{host: host, dataStore: dataStore, router: router, outputQueue: outputQueue, datasetAPI: datasetAPI, preview: preview}
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthcheck.Do)

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
	api.router.HandleFunc("/filter-outputs/{filter_output_id}/preview", api.getFilterOutputPreview).Methods("GET")

	if enablePrivateEndpoints {
		api.router.HandleFunc("/filter-outputs/{filter_output_id}", api.updateFilterOutput).Methods("PUT")
	}

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
