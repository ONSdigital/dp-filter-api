package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
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
	logData := log.Data{"filter_blueprint_id": filterBlueprintID}
	log.InfoCtx(r.Context(), "getting filter blueprint dimensions", logData)

	auditParams := common.Params{"filter_blueprint_id": filterBlueprintID}
	if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getDimensionsAction, actionAttempted, w, auditErr, logData)
		return
	}

	filter, err := api.getFilterBlueprint(r.Context(), filterBlueprintID)
	if err != nil {
		log.ErrorC("unable to get dimensions for filter blueprint", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	logData["dimensions"] = filter.Dimensions

	if len(filter.Dimensions) == 0 {
		log.ErrorCtx(r.Context(), filters.ErrDimensionNotFound, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	b, err := json.Marshal(filter.Dimensions)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint dimensions into bytes", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getDimensionsAction, actionSuccessful, w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		log.ErrorC("failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.InfoCtx(r.Context(), "got dimensions for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	log.InfoCtx(r.Context(), "getting filter blueprint dimension", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	if auditErr := api.auditor.Record(r.Context(), getDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	if _, err := api.getFilterBlueprint(r.Context(), filterBlueprintID); err != nil {
		log.ErrorCtx(r.Context(), err, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		if err == filters.ErrFilterBlueprintNotFound {
			setErrorCode(w, err, statusBadRequest)
			return
		}
		setErrorCode(w, err)
		return
	}

	if err := api.dataStore.GetFilterDimension(filterBlueprintID, name); err != nil {
		log.ErrorCtx(r.Context(), err, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getDimensionAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getDimensionAction, actionSuccessful, w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.InfoCtx(r.Context(), "got filtered blueprint dimension", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	log.InfoCtx(r.Context(), "removing filter blueprint dimension", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(r.Context(), removeDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), removeDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	if err := api.removeFilterBlueprintDimension(r.Context(), filterBlueprintID, dimensionName); err != nil {
		log.ErrorC("failed to remove dimension from filter blueprint", err, logData)
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
	w.WriteHeader(http.StatusOK)

	log.InfoCtx(r.Context(), "delete dimension from filter blueprint", logData)
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
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	log.InfoCtx(r.Context(), "add filter blueprint dimension", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(r.Context(), addDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), addDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		if auditErr := api.auditor.Record(r.Context(), addDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), addDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	options = removeDuplicateOptions(options)

	err = api.addFilterBlueprintDimension(r.Context(), filterBlueprintID, dimensionName, options)
	if err != nil {
		log.ErrorCtx(r.Context(), err, logData)
		if auditErr := api.auditor.Record(r.Context(), addDimensionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), addDimensionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}

		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}

		setErrorCode(w, err)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), addDimensionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(r.Context(), addDimensionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.InfoCtx(r.Context(), "created new dimension for filter blueprint", logData)
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
	log.InfoCtx(ctx, "check filter dimensions and dimension options before calling api, see version number", logData)

	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.getDimensions(ctx, dataset)
	if err != nil {
		log.ErrorC("failed to retrieve a list of dimensions from the dataset API", err, logData)
		return err
	}

	dimension := models.Dimension{
		Name:    name,
		Options: options,
	}

	if err = models.ValidateFilterDimensions([]models.Dimension{dimension}, datasetDimensions); err != nil {
		log.ErrorC("filter dimensions failed validation", err, logData)
		return err
	}

	// Call dimension options endpoint
	datasetDimensionOptions, err := api.getDimensionOptions(ctx, dataset, dimension.Name)
	if err != nil {
		log.ErrorC("failed to retrieve a list of dimension options from the dataset API", err, logData)
		return err
	}

	var incorrectDimensionOptions []string
	incorrectOptions := models.ValidateFilterDimensionOptions(dimension.Options, datasetDimensionOptions)
	if incorrectOptions != nil {
		incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
	}

	if incorrectDimensionOptions != nil {
		err = fmt.Errorf("incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.ErrorC("incorrect dimension options chosen", err, logData)
		return err
	}

	return nil
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
