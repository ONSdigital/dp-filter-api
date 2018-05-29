package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	// audit actions
	getDimensionsAction = "getFilterBlueprintDimensionsHandler"
	getDimensionAction  = "getFilterBlueprintDimensionHandler"
)

func (api *FilterAPI) getFilterBlueprintDimensionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterBlueprintID}
	log.Info("getting filter blueprint dimensions", logData)

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
		log.Error(filters.ErrDimensionNotFound, logData)
		if auditErr := api.auditor.Record(r.Context(), getDimensionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getDimensionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	bytes, err := json.Marshal(filter.Dimensions)
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
	_, err = w.Write(bytes)
	if err != nil {
		log.ErrorC("failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info("got dimensions for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	log.Info("getting filter blueprint dimension", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           name,
	}
	if auditErr := api.auditor.Record(r.Context(), getDimensionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getDimensionAction, actionAttempted, w, auditErr, logData)
		return
	}

	if _, err := api.getFilterBlueprint(r.Context(), filterBlueprintID); err != nil {
		log.Error(err, logData)
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
		log.Error(err, logData)
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

	log.Info("got filtered blueprint dimension", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("removing filter blueprint dimension", logData)

	filter, err := api.getFilterBlueprint(r.Context(), filterID)
	if err != nil {
		log.Error(err, logData)
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		default:
			setErrorCode(w, err)
		}
		return
	}

	if filter.State == models.SubmittedState {
		log.Error(errForbidden, logData)
		setErrorCode(w, errForbidden)
		return
	}

	if err := api.dataStore.RemoveFilterDimension(filterID, name); err != nil {
		log.ErrorC("unable to remove dimension from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete dimension from filter blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("add filter blueprint dimension", logData)

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	filterBlueprint, err := api.getFilterBlueprint(r.Context(), filterID)
	if err != nil {
		log.Error(err, logData)
		setErrorCode(w, err)
		return
	}
	logData["current_filter_blueprint"] = filterBlueprint

	if filterBlueprint.State == models.SubmittedState {
		log.Error(errForbidden, logData)
		setErrorCode(w, errForbidden)
		return
	}

	if err = api.checkNewFilterDimension(r.Context(), name, options, *filterBlueprint.Dataset); err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		if err == filters.ErrVersionNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = api.dataStore.AddFilterDimension(filterID, name, options, filterBlueprint.Dimensions); err != nil {
		log.ErrorC("failed to add dimension to filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension for filter blueprint", logData)
}

func (api *FilterAPI) checkNewFilterDimension(ctx context.Context, name string, options []string, dataset models.Dataset) error {
	logData := log.Data{"dimension_name": name, "dimension_options": options, "dataset": dataset}
	log.Info("check filter dimensions and dimension options before calling api, see version number", logData)

	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.datasetAPI.GetVersionDimensions(ctx, dataset)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimensions from the dataset API", err, logData)
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
	datasetDimensionOptions, err := api.datasetAPI.GetVersionDimensionOptions(ctx, dataset, dimension.Name)
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
