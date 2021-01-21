package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	datasetAPI "github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/utils"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *FilterAPI) getFilterBlueprintDimensionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterBlueprintID}
	ctx := r.Context()
	log.Event(ctx, "getting filter blueprint dimensions", log.INFO, logData)

	// get limit from query parameters, or default value
	limit, err := getPositiveIntQueryParameter(r.URL.Query(), "limit", api.defaultLimit)
	if err != nil {
		log.Event(ctx, "failed to obtain limit from request query parameters", log.ERROR, logData)
		setErrorCode(w, err)
		return
	}

	// get offset from query parameters, or default value
	offset, err := getPositiveIntQueryParameter(r.URL.Query(), "offset", api.defaultOffset)
	if err != nil {
		log.Event(ctx, "failed to obtain offset from request query parameters", log.ERROR, logData)
		setErrorCode(w, err)
		return
	}

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Event(ctx, "unable to get dimensions for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["dimensions"] = filter.Dimensions

	if len(filter.Dimensions) == 0 {
		log.Event(ctx, "dimension not found", log.ERROR, log.Error(filters.ErrDimensionNotFound), logData)
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	var dimensionNames []string

	for _, dimension := range filter.Dimensions {
		dimensionNames = append(dimensionNames, dimension.Name)
	}

	sort.Strings(dimensionNames)
	slicedDimensionNames := slice(dimensionNames, offset, limit)

	var filterDimensions []models.Dimension

	for _, dimensionName := range slicedDimensionNames {
		for _, dimension := range filter.Dimensions {
			if dimension.Name == dimensionName {
				filterDimensions = append(filterDimensions, dimension)
				break
			}
		}
	}

	items := createPublicDimensions(filterDimensions, api.host, filter.FilterID)
	publicDimensions := models.PublicDimensions{
		Items:      items,
		Count:      len(items),
		TotalCount: len(filter.Dimensions),
		Offset:     offset,
		Limit:      limit,
	}
	b, err := json.Marshal(publicDimensions)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "got dimensions for filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	ctx := r.Context()
	log.Event(ctx, "getting filter blueprint dimension", log.INFO, logData)

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Event(ctx, "error getting filter blueprint", log.ERROR, log.Error(err), logData)
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(filterBlueprintID, name, filter.ETag)
	if err != nil {
		log.Event(ctx, "error getting filter dimension", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	publicDimension := createPublicDimension(*dimension, api.host, filterBlueprintID)
	b, err := json.Marshal(publicDimension)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "got filtered blueprint dimension", log.INFO, logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	ctx := r.Context()
	log.Event(ctx, "removing filter blueprint dimension", log.INFO, logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Event(ctx, "missing header", log.ERROR, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newETag, err := api.removeFilterBlueprintDimension(r.Context(), filterBlueprintID, dimensionName, eTag)
	if err != nil {
		log.Event(r.Context(), "failed to remove dimension from filter blueprint", log.ERROR, log.Error(err), logData)
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newETag)
	w.WriteHeader(http.StatusNoContent)

	log.Event(r.Context(), "delete dimension from filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) removeFilterBlueprintDimension(ctx context.Context, filterBlueprintID, dimensionName, eTag string) (newETag string, err error) {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		return "", err
	}

	var dimensionExists bool
	for _, dimension := range filter.Dimensions {
		if dimension.Name == dimensionName {
			dimensionExists = true
			break
		}
	}
	if !dimensionExists {
		return "", filters.ErrDimensionNotFound
	}

	return api.dataStore.RemoveFilterDimension(filterBlueprintID, dimensionName, filter.UniqueTimestamp, eTag, filter)
}

func (api *FilterAPI) addFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	ctx := r.Context()
	log.Event(ctx, "add filter blueprint dimension", log.INFO, logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Event(ctx, "missing header", log.ERROR, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Event(ctx, "unable to unmarshal request body", log.ERROR, log.Error(err), logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	options = removeDuplicateAndEmptyOptions(options)

	newETag, err := api.addFilterBlueprintDimension(ctx, filterBlueprintID, dimensionName, options, eTag)
	if err != nil {
		log.Event(ctx, "error adding filter blueprint dimension", log.ERROR, log.Error(err), logData)
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}
		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(filterBlueprintID, dimensionName, newETag)
	if err != nil {
		log.Event(ctx, "error getting filter dimension", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
	publicDimension := createPublicDimension(*dimension, api.host, filterBlueprintID)
	b, err := json.Marshal(publicDimension)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, newETag)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "created new dimension for filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) addFilterBlueprintDimension(ctx context.Context, filterBlueprintID, dimensionName string, options []string, eTag string) (newETag string, err error) {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID, eTag)
	if err != nil {
		return "", err
	}

	timestamp := filterBlueprint.UniqueTimestamp

	if err = api.checkNewFilterDimension(ctx, dimensionName, options, filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			return "", err
		}
		return "", filters.NewBadRequestErr(err.Error())
	}

	return api.dataStore.AddFilterDimension(filterBlueprintID, dimensionName, options, filterBlueprint.Dimensions, timestamp, eTag, filterBlueprint)
}

// checkNewFilterDimension validates that the dimension with the provided name is valid, by calling GetDimensions in Dataset API.
// Once the dimension name is validated, it validates the provided options array.
func (api *FilterAPI) checkNewFilterDimension(ctx context.Context, name string, options []string, dataset *models.Dataset) error {

	logData := log.Data{"dimension_name": name, "dimension_options": options, "dataset": dataset}
	log.Event(ctx, "check filter dimensions and dimension options before calling api, see version number", log.INFO, logData)

	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.getDimensions(ctx, dataset)
	if err != nil {
		log.Event(ctx, "failed to retrieve a list of dimensions from the dataset API", log.ERROR, log.Error(err), logData)
		return err
	}

	dimension := models.Dimension{
		Name:    name,
		Options: options,
	}

	if err = models.ValidateFilterDimensions([]models.Dimension{dimension}, datasetDimensions); err != nil {
		log.Event(ctx, "filter dimensions failed validation", log.ERROR, log.Error(err), logData)
		return err
	}

	return api.checkNewFilterDimensionOptions(ctx, dimension, dataset, logData)
}

// checkNewFilterDimensionOptions, assuming a valid dimension, this method checks that the options provided in the dimension struct are valid
// by calling getDimensionOptions in batches and verifying that the provided dataset contains all the provided dimension options.
func (api *FilterAPI) checkNewFilterDimensionOptions(ctx context.Context, dimension models.Dimension, dataset *models.Dataset, logData log.Data) error {
	logData["dimension"] = dimension
	maxLogOptions := utils.Min(30, api.maxDatasetOptions)

	// create map of all options that need to be found
	optionsNotFound := utils.CreateMap(dimension.Options)
	isFirstBatch := true

	// process batch func
	processBatch := func(batch datasetAPI.Options) (forceAbort bool, err error) {

		// (first iteration only) - logData
		if isFirstBatch {
			isFirstBatch = false
			logData["dimension_options_total"] = batch.TotalCount
			if batch.TotalCount > maxLogOptions {
				if batch.Items != nil && len(batch.Items) > 0 {
					logData["dimension_options_first"] = batch.Items[0]
				}
			} else {
				logData["dimension_options"] = batch
			}
		}

		// remove found items from notFound map
		for _, opt := range batch.Items {
			if _, found := optionsNotFound[opt.Option]; found {
				delete(optionsNotFound, opt.Option)
			}
		}

		// abort if all options have been found
		if len(optionsNotFound) == 0 {
			return true, nil
		}

		// otherwise continue with next batch
		return false, nil
	}

	// get dimension options from dataset API in batches
	if err := api.getDimensionOptionsBatchProcess(ctx, dimension, dataset, processBatch); err != nil {
		return err
	}

	log.Event(ctx, "dimension options successfully retrieved from dataset API", log.INFO, logData)

	// if there is any dimension that is not found, error
	if optionsNotFound != nil && len(optionsNotFound) > 0 {
		incorrectDimensionOptions := utils.CreateArray(optionsNotFound)
		logData["incorrect_dimension_options"] = incorrectDimensionOptions
		err := fmt.Errorf("incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.Event(ctx, "incorrect dimension options chosen", log.ERROR, log.Error(err), logData)
		return err
	}

	return nil
}

// createPublicDimensions wraps createPublicDimension for converting arrays of dimensions
func createPublicDimensions(inputDimensions []models.Dimension, host, filterID string) []*models.PublicDimension {

	var outputDimensions []*models.PublicDimension
	for _, inputDimension := range inputDimensions {

		publicDimension := createPublicDimension(inputDimension, host, filterID)
		outputDimensions = append(outputDimensions, publicDimension)
	}

	return outputDimensions
}

// createPublicDimension creates a PublicDimension struct from a Dimension struct
func createPublicDimension(dimension models.Dimension, host, filterID string) *models.PublicDimension {

	// split out filterID and URL from dimension.URL
	filterURL := fmt.Sprintf("%s/filters/%s", host, filterID)
	dimensionURL := fmt.Sprintf("%s/dimensions/%s", filterURL, dimension.Name)

	publicDim := &models.PublicDimension{
		Name: dimension.Name,
		Links: &models.PublicDimensionLinkMap{
			Self:    models.LinkObject{HRef: dimensionURL, ID: dimension.Name},
			Filter:  models.LinkObject{HRef: filterURL, ID: filterID},
			Options: models.LinkObject{HRef: dimensionURL + "/options"},
		},
	}
	return publicDim
}

func removeDuplicateAndEmptyOptions(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			if elements[v] != "" {
				result = append(result, elements[v])
			}
		}
	}

	return result
}
