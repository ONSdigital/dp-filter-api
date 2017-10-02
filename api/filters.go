package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
)

var (
	internalError = "Failed to process the request due to an internal error"
	badRequest    = "Bad request - Invalid request body"
	unauthorised  = "Unauthorised, request lacks valid authentication credentials"
	forbidden     = "Forbidden, the filter job has been locked as it has been submitted to be processed"
)

const internalToken = "internal_token"

func (api *FilterAPI) addFilterJob(w http.ResponseWriter, r *http.Request) {
	newFilter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// Create unique id
	newFilter.FilterID = (uuid.NewV4()).String()

	if err := newFilter.Validate(); err != nil {
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
	newFilter.Links = models.LinkMap{Version: instance.Links.Version}

	filterJob, err := api.dataStore.AddFilter(api.host, newFilter)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	// Remove new filter job dimensions and build dimension url
	filterJob.Dimensions = nil
	filterJob.DimensionListURL = "/filters/" + filterJob.FilterID + "/dimensions"

	bytes, err := json.Marshal(filterJob)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	if filterJob.State == "submitted" {

		api.jobQueue.Queue(&filterJob)

		log.Info("filter job message sent to kafka", log.Data{"filter_job_id": filterJob.FilterID})
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_job": filterJob})
		setErrorCode(w, err)
		return
	}

	log.Info("created new filter job", log.Data{"filter_job": filterJob})
}

func (api *FilterAPI) getFilterJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]

	filterJob, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(filterJob)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_job": filterJob})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered job", log.Data{"filter_id": filterID, "filter_job": filterJob})
}

func (api *FilterAPI) getFilterJobDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]

	dimensions, err := api.dataStore.GetFilterDimensions(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensions)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimensions": dimensions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered job", log.Data{"filter_job_id": filterID, "dimensions": dimensions})
}

func (api *FilterAPI) getFilterJobDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]
	name := vars["name"]

	err := api.dataStore.GetFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered job", log.Data{"filter_job_id": filterID, "dimension": name})
}

func (api *FilterAPI) removeFilterJobDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]
	name := vars["name"]

	err := api.dataStore.RemoveFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered job", log.Data{"filter_job_id": filterID, "dimension": name})
}

func (api *FilterAPI) addFilterJobDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	addDimension := &models.AddDimension{
		FilterID: vars["filter_job_id"],
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

	log.Info("created new dimension for filter job", log.Data{"dimension": addDimension})
}

func (api *FilterAPI) getFilterJobDimensionOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]
	name := vars["name"]

	dimensionOptions, err := api.dataStore.GetFilterDimensionOptions(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensionOptions)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension_options": dimensionOptions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension_options": dimensionOptions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered job", log.Data{"filter_job_id": filterID, "dimension_options": dimensionOptions})
}

func (api *FilterAPI) getFilterJobDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.GetFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension": name, "option": option})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered job", log.Data{"filter_job_id": filterID, "dimension": name, "option": option})
}

func (api *FilterAPI) addFilterJobDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	addDimensionOption := &models.AddDimensionOption{
		FilterID: vars["filter_job_id"],
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

	log.Info("created new dimension option for filter job", log.Data{"dimension_option": addDimensionOption})
}

func (api *FilterAPI) updateFilterJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]

	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID})
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	authenticationHeader := r.Header.Get(internalToken)

	var isAuthenticated bool
	if authenticationHeader != "" {
		isAuthenticated, err = api.checkAuthentication(authenticationHeader)
		if err != nil {
			log.Error(err, log.Data{"filter_job_id": filterID})
			setErrorCode(w, err)
			return
		}
	}

	filter.FilterID = filterID

	err = api.dataStore.UpdateFilter(isAuthenticated, filter)
	if err != nil {
		log.Error(err, log.Data{"filter": filter, "filter_job_id": filterID})
		setErrorCode(w, err)
		return
	}

	log.Info("filter updated", log.Data{"filter_job_id": filterID, "filter": filter})
	if filter.State == "submitted" {

		api.jobQueue.Queue(filter)
		log.Info("filter job message sent to kafka", log.Data{"filter_job_id": filterID})
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
}

func (api *FilterAPI) removeFilterJobDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_job_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.RemoveFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered job", log.Data{"filter_job_id": filterID, "dimension": name})
}

func (api *FilterAPI) checkAuthentication(header string) (bool, error) {
	if header != api.internalToken {
		authorisationError := errors.New("Not authorised")
		return false, authorisationError
	}

	return true, nil
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setErrorCode(w http.ResponseWriter, err error) {
	log.Debug("error is", log.Data{"error": err})
	switch {
	case err.Error() == "Not found":
		http.Error(w, "Filter job not found", http.StatusNotFound)
		return
	case err.Error() == "Dimension not found":
		http.Error(w, "Dimension not found", http.StatusNotFound)
		return
	case err.Error() == "Option not found":
		http.Error(w, "Option not found", http.StatusNotFound)
		return
	case err.Error() == "Bad request - filter job not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err.Error() == "Bad request - dimension not found":
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
