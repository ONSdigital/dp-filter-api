package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
)

var (
	internalError = "Internal server error"
	badRequest    = "Bad client request received"
)

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

	filterJob, err := api.dataStore.AddFilter(api.host, newFilter)
	if err != nil {
		log.Error(err, log.Data{"new-filter": newFilter})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(filterJob)
	if err != nil {
		log.Error(err, log.Data{"new-filter": newFilter})
		setErrorCode(w, err)
		return
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

func (api *FilterAPI) updateFilterJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filterId"]
	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	filter.FilterID = filterID

	err = api.dataStore.UpdateFilter(filterID, filter)
	if err != nil {
		log.Error(err, log.Data{"filter": filter, "filter_id": filterID})
		setErrorCode(w, err)
		return
	}

	log.Info("filter updated", log.Data{"filter_id": filterID, "filter": filter})

	if filter.State == "submitted" {

		api.jobQueue.Queue(filter)

		log.Info("filter job message sent to kafka", log.Data{"filter_id": filterID})
	}
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
	case err.Error() == "Bad request":
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	case err.Error() == "Forbidden":
		http.Error(w, "Filter job request forbidden", http.StatusForbidden)
		return
	case err.Error() == "Not authorised":
		http.Error(w, "Filter job request not authorised", http.StatusUnauthorized)
		return
	case err != nil:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
