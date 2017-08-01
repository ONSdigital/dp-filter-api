package api

import "github.com/gorilla/mux"

// FilterAPI manages importing filters against a dataset
type FilterAPI struct {
	host      string
	dataStore DataStore
	router    *mux.Router
}

// CreateFilterAPI manages all the routes configured to API
func CreateFilterAPI(host string, router *mux.Router, dataStore DataStore) *FilterAPI {
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthCheck)

	api := FilterAPI{host: host, dataStore: dataStore, router: router}
	api.router.HandleFunc("/filters", api.addFilter).Methods("POST")
	return &api
}
