package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
)

var internalError = "Internal server error"

func (api *FilterAPI) addFilter(w http.ResponseWriter, r *http.Request) {
	newFilter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}

	if err := newFilter.Validate(); err != nil {
		log.Error(err, log.Data{})
		http.Error(w, "Bad client request received", http.StatusBadRequest)
		return
	}

	filterJob, err := api.dataStore.AddFilter(api.host, newFilter)
	if err != nil {
		log.Error(err, log.Data{"new-filter": newFilter})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(filterJob)
	if err != nil {
		log.Error(err, log.Data{"new-filter": newFilter})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_job": filterJob})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	log.Info("created new filter job", log.Data{"filter_job": filterJob})
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
