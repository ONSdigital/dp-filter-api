package api

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

//go:generate moq -out datastoretest/preview.go -pkg datastoretest . PreviewDataset

var httpServer *server.Server

// DatasetAPI - An interface used to access the DatasetAPI
type DatasetAPI interface {
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string) (m dataset.Options, err error)
}

// OutputQueue - An interface used to queue filter outputs
type OutputQueue interface {
	Queue(output *models.Filter) error
}

// PreviewDataset An interface used to generate previews
type PreviewDataset interface {
	GetPreview(ctx context.Context, filter *models.Filter, limit int) (*models.FilterPreview, error)
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host                 string
	dataStore            DataStore
	outputQueue          OutputQueue
	router               *mux.Router
	datasetAPI           DatasetAPI
	preview              PreviewDataset
	downloadServiceURL   string
	downloadServiceToken string
	auditor              audit.AuditorService
	serviceAuthToken     string
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(ctx context.Context,
	host, bindAddr, zebedeeURL string,
	datastore DataStore,
	outputQueue OutputQueue,
	errorChan chan error,
	datasetAPI DatasetAPI,
	preview PreviewDataset,
	enablePrivateEndpoints bool,
	downloadServiceURL,
	downloadServiceToken string,
	auditor audit.AuditorService,
	hc *healthcheck.HealthCheck) {

	router := mux.NewRouter()
	routes(host,
		router,
		datastore,
		outputQueue,
		datasetAPI,
		preview,
		enablePrivateEndpoints,
		downloadServiceURL,
		downloadServiceToken,
		auditor)

	middlewareChain := alice.New(
		middleware.Whitelist(middleware.HealthcheckFilter(hc.Handler)),
		collectionID.CheckHeader)

	if enablePrivateEndpoints {
		log.Event(ctx, "private endpoints are enabled. using identity middleware", log.INFO)
		identityHandler := identity.Handler(zebedeeURL)
		middlewareChain = middlewareChain.Append(identityHandler)
	}

	alice := middlewareChain.Then(router)
	httpServer = server.New(bindAddr, alice)

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
			errorChan <- err
		}
	}()
}

// routes contain all endpoints for API
func routes(host string, router *mux.Router, dataStore DataStore, outputQueue OutputQueue, datasetAPI DatasetAPI, preview PreviewDataset, enablePrivateEndpoints bool, downloadServiceURL, downloadServiceToken string, auditor audit.AuditorService) *FilterAPI {

	api := FilterAPI{host: host,
		dataStore:            dataStore,
		router:               router,
		outputQueue:          outputQueue,
		datasetAPI:           datasetAPI,
		preview:              preview,
		downloadServiceURL:   downloadServiceURL,
		downloadServiceToken: downloadServiceToken,
		auditor:              auditor,
	}

	api.router.HandleFunc("/filters", api.postFilterBlueprintHandler).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.getFilterBlueprintHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.putFilterBlueprintHandler).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions", api.getFilterBlueprintDimensionsHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.getFilterBlueprintDimensionHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.addFilterBlueprintDimensionHandler).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.removeFilterBlueprintDimensionHandler).Methods("DELETE")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options", api.getFilterBlueprintDimensionOptionsHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.getFilterBlueprintDimensionOptionHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.addFilterBlueprintDimensionOptionHandler).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", api.removeFilterBlueprintDimensionOptionHandler).Methods("DELETE")

	api.router.HandleFunc("/filter-outputs/{filter_output_id}", api.getFilterOutputHandler).Methods("GET")
	api.router.HandleFunc("/filter-outputs/{filter_output_id}/preview", api.getFilterOutputPreviewHandler).Methods("GET")

	if enablePrivateEndpoints {
		api.router.HandleFunc("/filter-outputs/{filter_output_id}", api.updateFilterOutputHandler).Methods("PUT")
		api.router.HandleFunc("/filter-outputs/{filter_output_id}/events", api.addEventHandler).Methods("POST")
	}

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Event(ctx, "graceful shutdown of http server complete", log.INFO)
	return nil
}
