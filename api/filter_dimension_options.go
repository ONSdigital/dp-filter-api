package api

import (
	"encoding/json"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"net/http"

	"fmt"
	"github.com/ONSdigital/dp-filter-api/filters"
	"context"
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

	options, err := api.getFilterBlueprintDimensionOptions(r.Context(), filterBlueprintID, dimensionName)
	if err != nil {
		log.ErrorC("failed to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	logData["options"] = options

	bytes, err := json.Marshal(options)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint dimension options into bytes", err, logData)
		http.Error(w, internalError, http.StatusInternalServerError)
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
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
		"option":              option,
	}
	log.Info("get filter blueprint dimension option", logData)

	err := api.getFilterBlueprintDimensionOption(r.Context(), filterID, name, option)
	if err != nil {
		log.ErrorC("unable to get dimension option for filter blueprint", err, logData)
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		default:
			setErrorCode(w, err)
		}
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got dimension option for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) (error) {

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
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{"filter_id": filterID, "dimension_name": name, "dimension_option": option}

	err := api.addFilterBlueprintDimensionOption(r.Context(), filterID, name, option)
	if err != nil {
		log.Error(err, logData)
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

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) (error) {

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
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]

	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
		"option":              option,
	}
	log.Info("remove filter blueprint dimension option", logData)


	err := api.removeFilterBlueprintDimensionOption(r.Context(), filterID, name, option)
	if err != nil {
		log.Error(err, logData)
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

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete dimension option on filter blueprint", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) (error) {

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