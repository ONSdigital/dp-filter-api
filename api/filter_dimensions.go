package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	datasetAPI "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/utils"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-net/v2/links"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (api *FilterAPI) getFilterBlueprintDimensionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterBlueprintID}
	ctx := r.Context()

	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	log.Info(ctx, "getting filter blueprint dimensions", logData)

	offset := api.defaultOffset
	limit := api.defaultLimit
	var err error

	if offsetParameter != "" {
		logData["offset"] = offsetParameter
		offset, err = validatePositiveInt(offsetParameter)
		if err != nil {
			log.Error(ctx, "failed to obtain offset from request query parameters", err, logData)
			setErrorCode(w, err)
			return
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = validatePositiveInt(limitParameter)
		if err != nil {
			log.Error(ctx, "failed to obtain limit from request query parameters", err, logData)
			setErrorCode(w, err)
			return
		}
	}

	if limit > api.maxLimit {
		logData["max_limit"] = api.maxLimit
		err = filters.ErrInvalidQueryParameter
		log.Error(ctx, "limit is greater than the maximum allowed", err, logData)
		setErrorCode(w, err)
		return
	}

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "unable to get dimensions for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	logData["filter out put:"] = filter

	logData["dimensions"] = filter.Dimensions

	if len(filter.Dimensions) == 0 {
		log.Error(ctx, "dimension not found", filters.ErrDimensionNotFound, logData)
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

	items := CreatePublicDimensions(filterDimensions, api.host.String(), filter.FilterID)
	publicDimensions := models.PublicDimensions{
		Items:      items,
		Count:      len(items),
		TotalCount: len(filter.Dimensions),
		Offset:     offset,
		Limit:      limit,
	}
	logData["items out put:"] = items

	enableRewriting := true
	if enableRewriting {
		dimensionSearchAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)

		for i := range publicDimensions.Items {
			item := publicDimensions.Items[i].Links
			//self
			newSelfLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Self.HRef)
			fmt.Println("newSelfLink is: ", newSelfLink)
			if err == nil {
				publicDimensions.Items[i].Links.Self.HRef = newSelfLink
			}
			//filter
			newFilterLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Filter.HRef)
			fmt.Println("newFilterLink is: ", newFilterLink)
			if err == nil {
				publicDimensions.Items[i].Links.Filter.HRef = newFilterLink
			}
			//options
			newOptionsLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Options.HRef)
			fmt.Println("newOptionsLink is: ", newOptionsLink)
			if err == nil {
				publicDimensions.Items[i].Links.Options.HRef = newOptionsLink
			}
		}
	}

	b, err := json.Marshal(publicDimensions)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimensions into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "got dimensions for filter blueprint", logData)
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
	log.Info(ctx, "getting filter blueprint dimension", logData)

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "error getting filter blueprint", err, logData)
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(ctx, filterBlueprintID, name, filter.ETag)
	if err != nil {
		log.Error(ctx, "error getting filter dimension", err, logData)
		setErrorCode(w, err)
		return
	}

	publicDimension := CreatePublicDimension(*dimension, api.host.String(), filterBlueprintID)

	enableRewriting := true
	//testHost, _ := url.Parse("http//localabc:3333")
	if enableRewriting {
		dimensionSearchAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)

		//self
		newSelfLink, err := dimensionSearchAPILinksBuilder.BuildLink(publicDimension.Links.Self.HRef)
		fmt.Println("newSelfLink is: ", newSelfLink)
		if err == nil {
			publicDimension.Links.Self.HRef = newSelfLink
		}
		//filter
		newFilterLink, err := dimensionSearchAPILinksBuilder.BuildLink(publicDimension.Links.Filter.HRef)
		fmt.Println("newFilterLink is: ", newFilterLink)
		if err == nil {
			publicDimension.Links.Filter.HRef = newFilterLink
		}
		//options
		newOptionsLink, err := dimensionSearchAPILinksBuilder.BuildLink(publicDimension.Links.Options.HRef)
		fmt.Println("newOptionsLink is: ", newOptionsLink)
		if err == nil {
			publicDimension.Links.Options.HRef = newOptionsLink
		}
	}

	b, err := json.Marshal(publicDimension)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimensions into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "got filtered blueprint dimension", logData)
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
	log.Info(ctx, "removing filter blueprint dimension", logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newETag, err := api.removeFilterBlueprintDimension(r.Context(), filterBlueprintID, dimensionName, eTag)
	if err != nil {
		log.Error(r.Context(), "failed to remove dimension from filter blueprint", err, logData)
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

	log.Info(r.Context(), "delete dimension from filter blueprint", logData)
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

	return api.dataStore.RemoveFilterDimension(ctx, filterBlueprintID, dimensionName, filter.UniqueTimestamp, eTag, filter)
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
	log.Info(ctx, "add filter blueprint dimension", logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Error(ctx, "unable to unmarshal request body", err, logData)
		http.Error(w, BadRequest, http.StatusBadRequest)
		return
	}

	options = RemoveDuplicateAndEmptyOptions(options)

	newETag, err := api.addFilterBlueprintDimension(ctx, filterBlueprintID, dimensionName, options, eTag)
	if err != nil {
		log.Error(ctx, "error adding filter blueprint dimension", err, logData)
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}
		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(ctx, filterBlueprintID, dimensionName, newETag)
	if err != nil {
		log.Error(ctx, "error getting filter dimension", err, logData)
		setErrorCode(w, err)
		return
	}
	publicDimension := CreatePublicDimension(*dimension, api.host.String(), filterBlueprintID)
	b, err := json.Marshal(publicDimension)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimensions into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, newETag)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "created new dimension for filter blueprint", logData)
}

// Handler for a list of put operations against the filter dimensions
func (api *FilterAPI) putFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	ctx := r.Context()
	log.Info(ctx, "put filter blueprint dimension", logData)

	err := errors.New("filter not of type flexible")
	log.Error(ctx, "invalid filter type", err, logData)
	http.Error(w, BadRequest, http.StatusBadRequest)
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

	return api.dataStore.AddFilterDimension(ctx, filterBlueprintID, dimensionName, options, filterBlueprint.Dimensions, timestamp, eTag, filterBlueprint)
}

// checkNewFilterDimension validates that the dimension with the provided name is valid, by calling GetDimensions in Dataset API.
// Once the dimension name is validated, it validates the provided options array.
func (api *FilterAPI) checkNewFilterDimension(ctx context.Context, name string, options []string, dataset *models.Dataset) error {

	logData := log.Data{"dimension_name": name, "dimension_options": options, "dataset": dataset}
	log.Info(ctx, "check filter dimensions and dimension options before calling api, see version number", logData)

	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.getDimensions(ctx, dataset)
	if err != nil {
		log.Error(ctx, "failed to retrieve a list of dimensions from the dataset API", err, logData)
		return err
	}

	dimension := models.Dimension{
		Name:    name,
		Options: options,
	}

	if err = models.ValidateFilterDimensions([]models.Dimension{dimension}, datasetDimensions); err != nil {
		log.Error(ctx, "filter dimensions failed validation", err, logData)
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

	log.Info(ctx, "dimension options successfully retrieved from dataset API", logData)

	// if there is any dimension that is not found, error
	if optionsNotFound != nil && len(optionsNotFound) > 0 {
		incorrectDimensionOptions := utils.CreateArray(optionsNotFound)
		logData["incorrect_dimension_options"] = incorrectDimensionOptions
		err := fmt.Errorf("incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.Error(ctx, "incorrect dimension options chosen", err, logData)
		return err
	}

	return nil
}

// CreatePublicDimensions wraps CreatePublicDimension for converting arrays of dimensions
func CreatePublicDimensions(inputDimensions []models.Dimension, host, filterID string) []*models.PublicDimension {

	outputDimensions := make([]*models.PublicDimension, 0)
	for _, inputDimension := range inputDimensions {

		publicDimension := CreatePublicDimension(inputDimension, host, filterID)
		outputDimensions = append(outputDimensions, publicDimension)
	}

	return outputDimensions
}

// CreatePublicDimension creates a PublicDimension struct from a Dimension struct
func CreatePublicDimension(dimension models.Dimension, host, filterID string) *models.PublicDimension {

	// split out filterID and URL from dimension.URL
	filterURL := fmt.Sprintf("%s/filters/%s", host, filterID)
	dimensionURL := fmt.Sprintf("%s/dimensions/%s", filterURL, dimension.Name)

	publicDim := &models.PublicDimension{
		Name: dimension.Name,
		Links: &models.PublicDimensionLinkMap{
			Self:    &models.LinkObject{HRef: dimensionURL, ID: dimension.Name},
			Filter:  &models.LinkObject{HRef: filterURL, ID: filterID},
			Options: &models.LinkObject{HRef: dimensionURL + "/options"},
		},
	}
	return publicDim
}

func RemoveDuplicateAndEmptyOptions(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, v3 := range elements {
		if !encountered[v3] {
			encountered[v3] = true
			if v3 != "" {
				result = append(result, v3)
			}
		}
	}

	return result
}

func getStringArrayFromInterface(elements interface{}) ([]string, error) {
	result := []string{}

	v2, ok := elements.([]interface{})
	if !ok {
		return result, errors.New("Missing list of items")
	}

	for _, v := range v2 {
		v3, ok := v.(string)
		if !ok {
			return result, fmt.Errorf("Non string item in list, got: %v", v)
		}
		result = append(result, v3)
	}

	return result, nil
}
