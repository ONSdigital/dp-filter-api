package api

import "github.com/gorilla/mux"

// FilterAPI - A restful API used to manage importing filters against a dataset
type FilterAPI struct {
	host      string
	dataStore DataStore
	router    *mux.Router
}

// CreateFilterAPI - Create the api with all the routes configured
func CreateFilterAPI(host string, router *mux.Router, dataStore DataStore) *FilterAPI {
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthCheck)

	api := FilterAPI{host: host, dataStore: dataStore, router: router}
	api.router.HandleFunc("/filters", api.addFilter).Methods("POST")
	return &api
}
