package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/middleware"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-net/v2/responder"
	"github.com/gorilla/mux"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetAPI
//go:generate moq -out mock/filterflexapi.go -pkg mock . FilterFlexAPI

// DatasetAPI - An interface used to access the DatasetAPI
type DatasetAPI interface {
	Get(ctx context.Context, userToken, svcToken, collectionID, datasetID string) (dataset.DatasetDetails, error)
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptionsBatchProcess(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, optionIDs *[]string, processBatch dataset.OptionsBatchProcessor, batchSize, maxWorkers int) (err error)
}

type FilterFlexAPI interface {
	ForwardRequest(*http.Request) (*http.Response, error)
}

// OutputQueue - An interface used to queue filter outputs
type OutputQueue interface {
	Queue(output *models.Filter) error
}

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host                 *url.URL
	DatasetAPIURL        *url.URL
	maxRequestOptions    int
	Router               *mux.Router
	dataStore            DataStore
	outputQueue          OutputQueue
	datasetAPI           DatasetAPI
	FilterFlexAPI        FilterFlexAPI
	downloadServiceURL   string
	downloadServiceToken string
	serviceAuthToken     string
	defaultLimit         int
	maxLimit             int
	defaultOffset        int
	maxDatasetOptions    int
	BatchMaxWorkers      int
	enableURLRewriting   bool
}

// Setup manages all the routes configured to API
func Setup(
	cfg *config.Config,
	router *mux.Router,
	dataStore DataStore,
	outputQueue OutputQueue,
	datasetAPI DatasetAPI,
	filterFlexAPI FilterFlexAPI,
	hostURL *url.URL,
	datasetAPIURL *url.URL,
	enableURLRewriting bool) *FilterAPI {
	api := &FilterAPI{
		host:                 hostURL,
		DatasetAPIURL:        datasetAPIURL,
		maxRequestOptions:    cfg.MaxRequestOptions,
		Router:               router,
		dataStore:            dataStore,
		outputQueue:          outputQueue,
		datasetAPI:           datasetAPI,
		downloadServiceURL:   cfg.DownloadServiceURL,
		downloadServiceToken: cfg.DownloadServiceSecretKey,
		serviceAuthToken:     cfg.ServiceAuthToken,
		defaultLimit:         cfg.MongoConfig.Limit,
		maxLimit:             cfg.DefaultMaxLimit,
		defaultOffset:        cfg.MongoConfig.Offset,
		maxDatasetOptions:    cfg.MaxDatasetOptions,
		BatchMaxWorkers:      cfg.BatchMaxWorkers,
		enableURLRewriting:   enableURLRewriting,
	}

	// middleware
	assert := middleware.NewAssert(
		responder.New(),
		datasetAPI,
		filterFlexAPI,
		dataStore,
		cfg.ServiceAuthToken,
		cfg.AssertDatasetType,
	)

	// routes
	api.Router.Handle("/filters", assert.DatasetType(http.HandlerFunc(api.postFilterBlueprintHandler))).Methods("POST")
	api.Router.Handle("/filters/{filter_blueprint_id}", assert.FilterType(http.HandlerFunc(api.getFilterBlueprintHandler))).Methods("GET")
	api.Router.Handle("/filters/{filter_blueprint_id}", assert.FilterType(http.HandlerFunc(api.putFilterBlueprintHandler))).Methods("PUT")
	api.Router.Handle("/filters/{filter_blueprint_id}/submit", assert.FilterType(http.HandlerFunc(api.postFilterBlueprintSubmitHandler))).Methods("POST")

	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions", assert.FilterType(http.HandlerFunc(api.getFilterBlueprintDimensionsHandler))).Methods("GET")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions", assert.FilterType(http.HandlerFunc(api.filterFlexNullEndpointHandler))).Methods("POST")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}", assert.FilterType(http.HandlerFunc(api.putFilterBlueprintDimensionHandler))).Methods("PUT")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}", assert.FilterType(http.HandlerFunc(api.getFilterBlueprintDimensionHandler))).Methods("GET")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}", assert.FilterType(http.HandlerFunc(api.addFilterBlueprintDimensionHandler))).Methods("POST")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}", assert.FilterType(http.HandlerFunc(api.patchFilterBlueprintDimensionHandler))).Methods("PATCH")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}", assert.FilterType(http.HandlerFunc(api.removeFilterBlueprintDimensionHandler))).Methods("DELETE")

	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", assert.FilterType(http.HandlerFunc(api.removeFilterBlueprintDimensionOptionHandler))).Methods("DELETE")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", assert.FilterType(http.HandlerFunc(api.addFilterBlueprintDimensionOptionHandler))).Methods("POST")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}/options/{option}", assert.FilterType(http.HandlerFunc(api.getFilterBlueprintDimensionOptionHandler))).Methods("GET")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}/options", assert.FilterType(http.HandlerFunc(api.getFilterBlueprintDimensionOptionsHandler))).Methods("GET")
	api.Router.Handle("/filters/{filter_blueprint_id}/dimensions/{name}/options", assert.FilterType(http.HandlerFunc(api.filterFlexNullEndpointHandler))).Methods("DELETE")

	api.Router.Handle("/filter-outputs/{filter_output_id}", assert.FilterOutputType(http.HandlerFunc(api.getFilterOutputHandler))).Methods("GET")
	api.Router.Handle("/filter-outputs/{filter_output_id}", assert.FilterOutputType(http.HandlerFunc(api.updateFilterOutputHandler))).Methods("PUT")

	if cfg.EnablePrivateEndpoints {
		api.Router.Handle("/filter-outputs/{filter_output_id}/events", assert.FilterOutputType(http.HandlerFunc(api.addEventHandler))).Methods("POST")
	}

	return api
}

// validatePositiveInt obtains the positive int value corresponding to the provided string
func validatePositiveInt(parameter string) (val int, err error) {
	val, err = strconv.Atoi(parameter)
	if err != nil {
		return -1, filters.ErrInvalidQueryParameter
	}
	if val < 0 {
		return 0, filters.ErrInvalidQueryParameter
	}
	return val, nil
}
