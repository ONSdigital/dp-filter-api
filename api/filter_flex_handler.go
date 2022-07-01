package api

import (
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

/*
   filterFlexNullEndpointHandler is used to log a Filter Flex
   request that does not have an equivalent
   CMD Journey.

   Used currently for DELETE filters/{id}/dimensions/{name}/options
*/
func (api *FilterAPI) filterFlexNullEndpointHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"attempted_url":       r.URL.String(),
	}

	ctx := r.Context()

	log.Error(ctx, "bad route", errors.New("attempted filter flex route with no CMD journey"), logData)
	http.Error(w, BadRequest, http.StatusBadRequest)
}
