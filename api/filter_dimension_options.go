package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"net/http"

	"context"
	"fmt"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/go-ns/common"
)

const (
	// audit actions
	getOptionsAction   = "getFilterBlueprintDimensionOptions"
	getOptionAction    = "getFilterBlueprintDimensionOption"
	removeOptionAction = "removeFilterBlueprintDimensionOption"
	addOptionAction    = "addFilterBlueprintDimensionOption"
)

func (api *FilterAPI) getFilterBlueprintDimensionOptionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	log.Info("get filter blueprint dimension options", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(r.Context(), getOptionsAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getOptionsAction, actionAttempted, w, auditErr, logData)
		return
	}

	options, err := api.getFilterBlueprintDimensionOptions(r.Context(), filterBlueprintID, dimensionName)
	if err != nil {
		log.ErrorC("failed to get dimension options for filter blueprint", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getOptionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getOptionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	logData["options"] = options

	bytes, err := json.Marshal(options)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint dimension options into bytes", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getOptionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getOptionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getOptionsAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getOptionsAction, actionSuccessful, w, auditErr, logData)
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

	log.Info("got dimension options for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprintID, dimensionName string) ([]models.DimensionOption, error) {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return nil, err
	}

	var options []models.DimensionOption
	dimensionFound := false
	for _, dimension := range filter.Dimensions {

		if dimension.Name == dimensionName {
			dimensionFound = true
			for _, option := range dimension.Options {
				url := fmt.Sprintf("%s/filter/%s/dimensions/%s/option/%s", api.host, filterBlueprintID, dimension.Name, option)
				dimensionOption := models.DimensionOption{Option: option, DimensionOptionURL: url}
				options = append(options, dimensionOption)
			}
		}
	}

	if !dimensionFound {
		return nil, filters.ErrDimensionNotFound
	}

	return options, nil
}

func (api *FilterAPI) getFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	log.Info("get filter blueprint dimension option", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	if auditErr := api.auditor.Record(r.Context(), getOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	err := api.getFilterBlueprintDimensionOption(r.Context(), filterBlueprintID, dimensionName, option)
	if err != nil {
		log.ErrorC("unable to get dimension option for filter blueprint", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		default:
			setErrorCode(w, err)
		}
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getOptionAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getOptionAction, actionSuccessful, w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got dimension option for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) error {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	optionFound := false
	dimensionFound := false
	for _, d := range filter.Dimensions {
		if d.Name == dimensionName {
			dimensionFound = true
			for _, o := range d.Options {
				if o == option {
					optionFound = true
				}
			}
		}
	}

	if !dimensionFound {
		return filters.ErrDimensionNotFound
	}

	if !optionFound {
		return filters.ErrOptionNotFound
	}

	return nil
}

func (api *FilterAPI) addFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]
	logData := log.Data{"filter_blueprint_id": filterBlueprintID, "dimension_name": dimensionName, "dimension_option": option}

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	if auditErr := api.auditor.Record(r.Context(), addOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), addOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	err := api.addFilterBlueprintDimensionOption(r.Context(), filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Error(err, logData)
		if auditErr := api.auditor.Record(r.Context(), addOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), addOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		case filters.ErrVersionNotFound:
			setErrorCode(w, err, statusUnprocessableEntity)
		default:
			setErrorCode(w, err)
		}
		return
	}

	if auditErr := api.auditor.Record(r.Context(), addOptionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(r.Context(), addOptionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) error {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	if filterBlueprint.State == models.SubmittedState {
		return err
	}

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	if err = api.checkNewFilterDimension(ctx, dimensionName, []string{option}, *filterBlueprint.Dataset); err != nil {
		if incorrectDimensionOptions.MatchString(err.Error()) {
			return filters.NewBadRequestErr(err.Error())
		}

		if incorrectDimension.MatchString(err.Error()) {
			return filters.NewBadRequestErr(err.Error())
		}

		return err
	}

	if err := api.dataStore.AddFilterDimensionOption(filterBlueprintID, dimensionName, option); err != nil {
		return err
	}

	return nil
}

func (api *FilterAPI) removeFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	log.Info("remove filter blueprint dimension option", logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	if auditErr := api.auditor.Record(r.Context(), removeOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), removeOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	err := api.removeFilterBlueprintDimensionOption(r.Context(), filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Error(err, logData)
		if auditErr := api.auditor.Record(r.Context(), removeOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), removeOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		case filters.ErrVersionNotFound:
			setErrorCode(w, err, statusUnprocessableEntity)
		default:
			setErrorCode(w, err)
		}
		return
	}

	if auditErr := api.auditor.Record(r.Context(), removeOptionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(r.Context(), removeOptionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete dimension option on filter blueprint", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) error {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	// Check if dimension exists
	var hasDimension bool
	for _, dimension := range filterBlueprint.Dimensions {
		if dimension.Name == dimensionName {
			hasDimension = true
			break
		}
	}

	if !hasDimension {
		return filters.ErrDimensionNotFound
	}

	if filterBlueprint.State == models.SubmittedState {
		return errForbidden
	}

	if err = api.dataStore.RemoveFilterDimensionOption(filterBlueprintID, dimensionName, option); err != nil {
		return err
	}

	return nil
}
