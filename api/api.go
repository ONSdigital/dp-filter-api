package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
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
func CreateFilterAPI(secretKey, host, bindAddr string, datastore DataStore, outputQueue OutputQueue, errorChan chan error, datasetAPI DatasetAPIer, preview PreviewDataset) {
	router := mux.NewRouter()
	routes(secretKey, host, router, datastore, outputQueue, datasetAPI, preview)

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
func routes(secretKey, host string, router *mux.Router, dataStore DataStore, outputQueue OutputQueue, datasetAPI DatasetAPIer, preview PreviewDataset) *FilterAPI {
	a := auth{key: secretKey}
	api := FilterAPI{host: host, dataStore: dataStore, router: router, outputQueue: outputQueue, datasetAPI: datasetAPI, preview: preview}
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthcheck.Do)

	api.router.HandleFunc("/filters", a.authenticate(api.addFilterBlueprint)).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", a.authenticate(api.getFilterBlueprint)).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", a.authenticate(api.updateFilterBlueprint)).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions", a.authenticate(api.getFilterBlueprintDimensions)).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", a.authenticate(api.getFilterBlueprintDimension)).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", a.authenticate(api.addFilterBlueprintDimension)).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", a.authenticate(api.removeFilterBlueprintDimension)).Methods("DELETE")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options", a.authenticate(api.getFilterBlueprintDimensionOptions)).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", a.authenticate(api.getFilterBlueprintDimensionOption)).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", a.authenticate(api.addFilterBlueprintDimensionOption)).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", a.authenticate(api.removeFilterBlueprintDimensionOption)).Methods("DELETE")

	api.router.HandleFunc("/filter-outputs/{filter_output_id}", a.authenticate(api.getFilterOutput)).Methods("GET")
	api.router.HandleFunc("/filter-outputs/{filter_output_id}", a.authenticate(api.updateFilterOutput)).Methods("PUT")
	api.router.HandleFunc("/filter-outputs/{filter_output_id}/preview", a.authenticate(api.getFilterOutputPreview)).Methods("GET")
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

//auth provides the key and functionality to check incoming requests for authentication
//and propogate authentication for further requests
type auth struct {
	key string
}

func (a *auth) authenticate(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Header.Get(string(internalTokenKey)) == a.key {
			ctx = context.WithValue(ctx, internalTokenKey, true)
		}
		handle(w, r.WithContext(ctx))
	})
}
