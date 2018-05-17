package api

import (
	"encoding/json"
	"net/http"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"fmt"
	"github.com/ONSdigital/dp-filter-api/filters"
)

func (api *FilterAPI) getFilterBlueprintDimensionOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("get filter blueprint dimension options", logData)

	filter, err := api.getFilter(r.Context(), filterID)
	if err != nil {
		log.ErrorC("unable to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	var options []models.DimensionOption
	dimensionFound := false
	for _, dimension := range filter.Dimensions {

		if dimension.Name == name {
			dimensionFound = true
			for _, option := range dimension.Options {
				url := fmt.Sprintf("%s/filter/%s/dimensions/%s/option/%s", api.host, filterID, dimension.Name, option)
				dimensionOption := models.DimensionOption{Option: option, DimensionOptionURL: url}
				options = append(options, dimensionOption)
			}

		}
	}

	if !dimensionFound {
		log.Error(filters.ErrDimensionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionNotFound)
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

func (api *FilterAPI) getFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
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

	filter, err := api.getFilter(r.Context(), filterID)
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

	dimensionFound := false
	optionFound := false
	for _, d := range filter.Dimensions {
		if d.Name == name {
			dimensionFound = true
			for _, o := range d.Options {
				if o == option {
					optionFound = true
				}
			}
		}
	}

	if !dimensionFound {
		log.Error(filters.ErrDimensionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	if !optionFound {
		log.Error(filters.ErrOptionNotFound, logData)
		setErrorCode(w, filters.ErrOptionNotFound)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got dimension option for filter blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{"filter_id": filterID, "dimension_name": name, "dimension_option": option}

	filterBlueprint, err := api.getFilter(r.Context(), filterID)
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

	if filterBlueprint.State == models.SubmittedState {
		log.Error(errForbidden, logData)
		setErrorCode(w, errForbidden, filterBlueprint.State)
		return
	}

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	if err = api.checkNewFilterDimension(r.Context(), name, []string{option}, *filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound {
			log.ErrorC("failed to select valid version", err, logData)
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}

		if incorrectDimensionOptions.MatchString(err.Error()) {
			log.ErrorC("failed to select valid filter dimension options", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if incorrectDimension.MatchString(err.Error()) {
			log.ErrorC("failed to select valid filter dimension", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.ErrorC("failed to successfully check filter dimensions", err, logData)
		setErrorCode(w, err)
		return
	}

	if err := api.dataStore.AddFilterDimensionOption(filterID, name, option); err != nil {
		log.ErrorC("failed to add dimension option to filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
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

	filterBlueprint, err := api.getFilter(r.Context(), filterID)
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

	// Check if dimension exists
	var hasDimension bool
	for _, dimension := range filterBlueprint.Dimensions {
		if dimension.Name == name {
			hasDimension = true
			break
		}
	}

	if !hasDimension {
		log.Error(filters.ErrDimensionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	if filterBlueprint.State == models.SubmittedState {
		log.Error(errForbidden, logData)
		setErrorCode(w, errForbidden)
		return
	}

	if err = api.dataStore.RemoveFilterDimensionOption(filterID, name, option); err != nil {
		log.ErrorC("unable to remove dimension option from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete dimension option on filter blueprint", logData)
}