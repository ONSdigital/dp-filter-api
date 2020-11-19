package api

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/gorilla/mux"
)

//go:generate moq -out datastoretest/preview.go -pkg datastoretest . PreviewDataset

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
	serviceAuthToken     string
}

// Setup manages all the routes configured to API
func Setup(host string,
	router *mux.Router,
	dataStore DataStore,
	outputQueue OutputQueue,
	datasetAPI DatasetAPI,
	preview PreviewDataset,
	enablePrivateEndpoints bool,
	downloadServiceURL, downloadServiceToken, serviceAuthToken string) *FilterAPI {

	api := &FilterAPI{
		host:                 host,
		dataStore:            dataStore,
		router:               router,
		outputQueue:          outputQueue,
		datasetAPI:           datasetAPI,
		preview:              preview,
		downloadServiceURL:   downloadServiceURL,
		downloadServiceToken: downloadServiceToken,
		serviceAuthToken:     serviceAuthToken,
	}

	api.router.HandleFunc("/filters", api.postFilterBlueprintHandler).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.getFilterBlueprintHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}", api.putFilterBlueprintHandler).Methods("PUT")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions", api.getFilterBlueprintDimensionsHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.getFilterBlueprintDimensionHandler).Methods("GET")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.addFilterBlueprintDimensionHandler).Methods("POST")
	api.router.HandleFunc("/filters/{filter_blueprint_id}/dimensions/{name}", api.patchFilterBlueprintDimensionHandler).Methods("PATCH")
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

	return api
}
