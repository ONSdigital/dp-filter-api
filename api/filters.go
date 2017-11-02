package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"fmt"

	uuid "github.com/satori/go.uuid"
)

var (
	internalError = "Failed to process the request due to an internal error"
	badRequest    = "Bad request - Invalid request body"
	unauthorised  = "Unauthorised, request lacks valid authentication credentials"
	forbidden     = "Forbidden, the filter output has been locked as it has been submitted to be processed"
)

const internalToken = "Internal-Token"

func (api *FilterAPI) addFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	submitted := r.FormValue("submitted")

	newFilter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// Create unique id
	newFilter.FilterID = uuid.NewV4().String()

	if err := newFilter.ValidateFilterBlueprint(); err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// add version information from datasetAPI
	instance, err := api.datasetAPI.GetInstance(r.Context(), newFilter.InstanceID)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	newFilter.Links = models.LinkMap{
		Dimensions: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s/dimensions", api.host, newFilter.FilterID),
		},
		Self: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s", api.host, newFilter.FilterID),
		},
		Version: instance.Links.Version,
	}

	filterBlueprint, err := api.dataStore.AddFilter(api.host, newFilter)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	// Remove new filter blueprint dimensions
	filterBlueprint.Dimensions = nil

	if submitted == "true" {

		// Create filter output resource and use id to pass into kafka
		filterOutput := api.createFilterOutputResource(filterBlueprint)

		filterBlueprint.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		filterBlueprint.Links.FilterOutput.ID = filterOutput.FilterID

		log.Info("filter output id sent in message to kafka", log.Data{"filter_output_id": filterOutput.FilterID})
	}

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filterBlueprint})
		setErrorCode(w, err)
		return
	}

	log.Info("created new filter blueprint", log.Data{"filter_blueprint": filterBlueprint})
}

func (api *FilterAPI) getFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]

	filterBlueprint, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		setErrorCode(w, err)
		return
	}

	filterBlueprint.Dimensions = nil

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filterBlueprint})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_id": filterID, "filter_blueprint": filterBlueprint})
}

func (api *FilterAPI) getFilterBlueprintDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]

	dimensions, err := api.dataStore.GetFilterDimensions(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensions)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimensions": dimensions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimensions": dimensions})
}

func (api *FilterAPI) getFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	err := api.dataStore.GetFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) removeFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	err := api.dataStore.RemoveFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) addFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	addDimension := &models.AddDimension{
		FilterID: vars["filter_blueprint_id"],
		Name:     vars["name"],
		Options:  options,
	}

	if err = api.dataStore.AddFilterDimension(addDimension); err != nil {
		log.Error(err, log.Data{"dimension": addDimension})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension for filter blueprint", log.Data{"dimension": addDimension})
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	dimensionOptions, err := api.dataStore.GetFilterDimensionOptions(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensionOptions)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.GetFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name, "option": option})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name, "option": option})
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	addDimensionOption := &models.AddDimensionOption{
		FilterID: vars["filter_blueprint_id"],
		Name:     vars["name"],
		Option:   vars["option"],
	}

	if err := api.dataStore.AddFilterDimensionOption(addDimensionOption); err != nil {
		log.Error(err, log.Data{"dimension_option": addDimensionOption})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", log.Data{"dimension_option": addDimensionOption})
}

func (api *FilterAPI) updateFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	var submitted string
	submitted = r.URL.Query().Get("submitted")

	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		// If filter blueprint has query parameter submitted set to true then
		// request can have an empty body
		if err != errors.New("Bad request - Missing data in body") && submitted != "true" {
			log.Error(err, log.Data{"filter_blueprint_id": filterID})
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
	}

	filter.FilterID = filterID

	currentFilter, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filter, "filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	newFilter := currentFilter
	if filter.InstanceID != "" {

		// add version information from datasetAPI for new instance
		instance, err := api.datasetAPI.GetInstance(r.Context(), newFilter.InstanceID)
		if err != nil {
			log.Error(err, log.Data{"new_filter": newFilter})
			setErrorCode(w, err)
			return
		}

		newFilter.InstanceID = filter.InstanceID
		newFilter.Links.Version.HRef = instance.Links.Version.HRef
	}

	if filter.Events.Error != nil {
		for _, errorEvent := range filter.Events.Error {
			newFilter.Events.Error = append(newFilter.Events.Error, errorEvent)
		}
	}

	if filter.Events.Info != nil {
		for _, infoEvent := range filter.Events.Info {
			newFilter.Events.Info = append(newFilter.Events.Info, infoEvent)
		}
	}

	err = api.dataStore.UpdateFilter(newFilter)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filter, "filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	if submitted == "true" {
		outputFilter := newFilter

		// Create filter output resource and use id to pass into kafka
		filterOutput := api.createFilterOutputResource(outputFilter)

		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID

		log.Info("filter output id sent in message to kafka", log.Data{"filter_output_id": filterOutput.FilterID})
	}

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": newFilter})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": newFilter})
		setErrorCode(w, err)
		return
	}

	log.Info("filter blueprint updated", log.Data{"filter_blueprint_id": filterID, "filter": newFilter})
}

func (api *FilterAPI) getFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]
	log.Info("got here", nil)

	filterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterOutputID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered output", log.Data{"filter_output_id": filterOutputID, "filter_output": filterOutput})
}

func (api *FilterAPI) updateFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	authenticationHeader := r.Header.Get(internalToken)

	isAuthenticated, err := api.checkAuthentication(authenticationHeader)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		setErrorCode(w, err)
		return
	}

	if !isAuthenticated {
		err := errors.New("Not authorised")
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		http.Error(w, unauthorised, http.StatusUnauthorized)
		return
	}

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err := filterOutput.ValidateFilterOutputUpdate(); err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	filterOutput.FilterID = filterOutputID

	// TODO check filter output resource for current downloads
	previousFilterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	updatedFilterOutput := checkFilterOutputQuery(previousFilterOutput, filterOutput)

	if err = api.dataStore.UpdateFilterOutput(updatedFilterOutput); err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("got filtered blueprint", log.Data{"filter_output_id": filterOutputID, "filter_output": filterOutput})
}

func (api *FilterAPI) createFilterOutputResource(newFilter *models.Filter) *models.Filter {
	filterOutput := newFilter
	filterOutput.FilterID = uuid.NewV4().String()
	filterOutput.State = "created"
	filterOutput.Links.Self.HRef = fmt.Sprintf("%s/filter-outputs/%s", api.host, filterOutput.FilterID)
	filterOutput.Links.Dimensions.HRef = ""
	filterOutput.Links.FilterBlueprint.HRef = newFilter.Links.Self.HRef
	filterOutput.Links.FilterBlueprint.ID = newFilter.FilterID

	// Clear out any event information to output document
	filterOutput.Events = models.Events{}

	// Downloads object should exist for filter output resource
	// even if it they are empty
	filterOutput.Downloads = &models.Downloads{
		CSV:  models.DownloadItem{},
		XLS:  models.DownloadItem{},
		JSON: models.DownloadItem{},
	}

	// Remove dimension url from output filter resource
	for i, _ := range newFilter.Dimensions {
		filterOutput.Dimensions[i].URL = ""
	}

	api.dataStore.CreateFilterOutput(filterOutput)

	api.jobQueue.Queue(filterOutput)

	return filterOutput
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.RemoveFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) checkAuthentication(header string) (bool, error) {
	if header != api.internalToken {
		authorisationError := errors.New("Not authorised")
		return false, authorisationError
	}

	return true, nil
}

func checkFilterOutputQuery(previousFilterOutput, filterOutput *models.Filter) *models.Filter {
	if previousFilterOutput.Downloads == nil {
		return filterOutput
	}

	if previousFilterOutput.Downloads.CSV.URL != "" {
		filterOutput.Downloads.CSV = previousFilterOutput.Downloads.CSV
	}

	if previousFilterOutput.Downloads.XLS.URL != "" {
		filterOutput.Downloads.XLS = previousFilterOutput.Downloads.XLS
	}

	if previousFilterOutput.Downloads.JSON.URL != "" {
		filterOutput.Downloads.JSON = previousFilterOutput.Downloads.JSON
	}

	return filterOutput
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setErrorCode(w http.ResponseWriter, err error) {
	log.Debug("error is", log.Data{"error": err})
	switch {
	case err.Error() == "Not found":
		http.Error(w, "Filter blueprint not found", http.StatusNotFound)
		return
	case err.Error() == "Dimension not found":
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err.Error() == "Option not found":
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err.Error() == "Filter output not found":
		http.Error(w, err.Error(), http.StatusNotFound)
	case err.Error() == "Instance not found":
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err.Error() == "Bad request - filter blueprint not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err.Error() == "Bad request - filter dimension not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err.Error() == "Bad request - filter or dimension not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err.Error() == "Bad request":
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	case err.Error() == "Forbidden":
		http.Error(w, forbidden, http.StatusForbidden)
		return
	case err.Error() == "Not authorised":
		http.Error(w, unauthorised, http.StatusUnauthorized)
		return
	case err != nil:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
