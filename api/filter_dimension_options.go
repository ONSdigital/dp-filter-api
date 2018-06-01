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

func (api *FilterAPI) getFilterBlueprintDimensionOptionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           dimensionName,
	}
	log.Info("get filter blueprint dimension options", logData)

	filter, err := api.getFilterBlueprint(r.Context(), filterID)
	if err != nil {
		log.ErrorC("unable to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	var options []models.DimensionOption
	dimensionFound := false
	for _, dimension := range filter.Dimensions {

		if dimension.Name == dimensionName {
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

	filter, err := api.getFilterBlueprint(r.Context(), filterID)
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
		log.Error(filters.ErrDimensionOptionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionOptionNotFound)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got dimension option for filter blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{"filter_id": filterID, "dimension_name": name, "dimension_option": option}

	filterBlueprint, err := api.getFilterBlueprint(r.Context(), filterID)
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

	timestamp := filterBlueprint.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	if err = api.checkNewFilterDimension(r.Context(), name, []string{option}, filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
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

	if err := api.dataStore.AddFilterDimensionOption(filterID, name, option, timestamp); err != nil {
		log.ErrorC("failed to add dimension option to filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", logData)
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

	filterBlueprint, err := api.getFilterBlueprint(r.Context(), filterID)
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

	timestamp := filterBlueprint.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	// Check if dimension and option exists
	var hasDimension bool
	var hasOption bool
	for _, dimension := range filterBlueprint.Dimensions {
		if dimension.Name == name {
			hasDimension = true
			for _, dimOption := range dimension.Options {
				if dimOption == option {
					hasOption = true
					break
				}
			}
			break
		}
	}

	if !hasDimension {
		log.Error(filters.ErrDimensionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionNotFound)
		return
	}

	if !hasOption {
		log.Error(filters.ErrDimensionOptionNotFound, logData)
		setErrorCode(w, filters.ErrDimensionOptionNotFound)
		return
	}

	if err = api.dataStore.RemoveFilterDimensionOption(filterID, name, option, timestamp); err != nil {
		log.ErrorC("unable to remove dimension option from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete dimension option on filter blueprint", logData)
}
