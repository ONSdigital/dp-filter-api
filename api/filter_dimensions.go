package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const (
	// audit actions
	getDimensionsAction   = "getFilterBlueprintDimensions"
	getDimensionAction    = "getFilterBlueprintDimension"
	removeDimensionAction = "removeFilterBlueprintDimension"
	addDimensionAction    = "addFilterBlueprintDimension"
)

func (api *FilterAPI) getFilterBlueprintDimensionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterBlueprintID, "action": getDimensionsAction}
	ctx := r.Context()
	log.Event(ctx, "getting filter blueprint dimensions", log.INFO, logData)

	auditParams := common.Params{"filter_blueprint_id": filterBlueprintID}
	if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getDimensionsAction, actionAttempted, w, auditErr, logData)
		return
	}

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "unable to get dimensions for filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	logData["dimensions"] = filter.Dimensions

	if len(filter.Dimensions) == 0 {
		log.Event(ctx, "dimension not found", log.ERROR, log.Error(filters.ErrDimensionNotFound), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	publicDimensions := createPublicDimensions(filter.Dimensions, api.host, filter.FilterID)
	b, err := json.Marshal(publicDimensions)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getDimensionsAction, actionSuccessful, w, auditErr, logData)
		return
	}

	setJSONContentType(w)
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
		"action":              getDimensionAction,
	}
	ctx := r.Context()
	log.Event(ctx, "getting filter blueprint dimension", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	if auditErr := api.auditor.Record(ctx, getDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	if _, err := api.getFilterBlueprint(ctx, filterBlueprintID); err != nil {
		log.Event(ctx, "error getting filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(filterBlueprintID, name)
	if err != nil {
		log.Event(ctx, "error getting filter dimension", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	publicDimension := createPublicDimension(*dimension, api.host, filterBlueprintID)
	b, err := json.Marshal(publicDimension)
	if err != nil {

		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, getDimensionAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getDimensionAction, actionSuccessful, w, auditErr, logData)
		return
	}

	setJSONContentType(w)
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
		"action":              removeDimensionAction,
	}
	log.Event(r.Context(), "removing filter blueprint dimension", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(r.Context(), removeDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), removeDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	if err := api.removeFilterBlueprintDimension(r.Context(), filterBlueprintID, dimensionName); err != nil {
		log.Event(r.Context(), "failed to remove dimension from filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(r.Context(), removeDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), removeDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), removeDimensionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(r.Context(), removeDimensionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Event(r.Context(), "delete dimension from filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) removeFilterBlueprintDimension(ctx context.Context, filterBlueprintID, dimensionName string) error {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	var dimensionExists bool
	for _, dimension := range filter.Dimensions {
		if dimension.Name == dimensionName {
			dimensionExists = true
			break
		}
	}
	if !dimensionExists {
		return filters.ErrDimensionNotFound
	}

	timestamp := filter.UniqueTimestamp

	return api.dataStore.RemoveFilterDimension(filterBlueprintID, dimensionName, timestamp)
}

func (api *FilterAPI) addFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {

	defer request.DrainBody(r)

	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"action":              addDimensionAction,
	}
	ctx := r.Context()
	log.Event(ctx, "add filter blueprint dimension", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(ctx, addDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, addDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Event(ctx, "unable to unmarshal request body", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, addDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, addDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	options = removeDuplicateOptions(options)

	err = api.addFilterBlueprintDimension(ctx, filterBlueprintID, dimensionName, options)
	if err != nil {
		log.Event(ctx, "error adding filter blueprint dimension", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, addDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, addDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}

		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}

		setErrorCode(w, err)
		return
	}

	dimension, err := api.dataStore.GetFilterDimension(filterBlueprintID, dimensionName)
	if err != nil {
		log.Event(ctx, "error getting filter dimension", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}
	publicDimension := createPublicDimension(*dimension, api.host, filterBlueprintID)
	b, err := json.Marshal(publicDimension)
	if err != nil {

		log.Event(ctx, "failed to marshal filter blueprint dimensions into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, addDimensionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(ctx, addDimensionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "created new dimension for filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) addFilterBlueprintDimension(ctx context.Context, filterBlueprintID, dimensionName string, options []string) error {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	timestamp := filterBlueprint.UniqueTimestamp

	if err = api.checkNewFilterDimension(ctx, dimensionName, options, filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			return err
		}
		return filters.NewBadRequestErr(err.Error())
	}

	return api.dataStore.AddFilterDimension(filterBlueprintID, dimensionName, options, filterBlueprint.Dimensions, timestamp)
}

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

	// Call dimension options endpoint
	datasetDimensionOptions, err := api.getDimensionOptions(ctx, dataset, dimension.Name)
	if err != nil {
		log.Event(ctx, "failed to retrieve a list of dimension options from the dataset API", log.ERROR, log.Error(err), logData)
		return err
	}

	var incorrectDimensionOptions []string
	incorrectOptions := models.ValidateFilterDimensionOptions(dimension.Options, datasetDimensionOptions)
	if incorrectOptions != nil {
		incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
	}

	if incorrectDimensionOptions != nil {
		err = fmt.Errorf("incorrect dimension options chosen: %v", incorrectDimensionOptions)
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

func removeDuplicateOptions(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}
