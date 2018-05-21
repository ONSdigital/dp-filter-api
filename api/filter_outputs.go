package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"fmt"

	"strconv"

	identity "github.com/ONSdigital/go-ns/common"
	"github.com/satori/go.uuid"
	"github.com/ONSdigital/dp-filter-api/filters"
)

var (
	errRequestLimitNotNumber = errors.New("requested limit is not a number")
	errMissingDimensions     = errors.New("missing dimensions")
)

func (api *FilterAPI) getFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.Info("getting filter output", logData)

	filterOutput, err := api.getOutput(r, filterOutputID)
	if err != nil {
		log.ErrorC("unable to get filter output", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["filter_output"] = filterOutput

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.ErrorC("failed to marshal filter output into bytes", err, logData)
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

	log.Info("got filter output", logData)
}

func (api *FilterAPI) updateFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.Info("updating filter output", logData)

	if !identity.IsCallerPresent(r.Context()) {
		log.ErrorC("failed to update filter output", filters.ErrUnauthorised, logData)
		setErrorCode(w, filters.ErrUnauthorised)
		return
	}

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}
	logData["filter_output"] = filterOutput

	// check filter output resource for current downloads and published flag
	previousFilterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.ErrorC("unable to get current filter output", err, logData)
		setErrorCode(w, err)
		return
	}

	if err = filterOutput.ValidateFilterOutputUpdate(previousFilterOutput); err != nil {
		log.ErrorC("filter output failed validation", err, logData)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	filterOutput.FilterID = filterOutputID

	// Set the published flag to the value currently stored on filter output resources
	// unless the request contains an update to the flag
	if previousFilterOutput.Published != nil && *previousFilterOutput.Published == models.Published {
		filterOutput.Published = &models.Published
	}

	filterOutputUpdate := buildDownloadsObject(previousFilterOutput, filterOutput, api.downloadServiceURL)

	if err = api.dataStore.UpdateFilterOutput(filterOutputUpdate); err != nil {
		log.ErrorC("unable to update filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("update filter output", logData)
}

func (api *FilterAPI) createFilterOutputResource(newFilter *models.Filter, filterBlueprintID string) (models.Filter, error) {
	filterOutput := *newFilter
	filterOutput.FilterID = uuid.NewV4().String()
	filterOutput.State = models.CreatedState
	filterOutput.Links.Self.HRef = fmt.Sprintf("%s/filter-outputs/%s", api.host, filterOutput.FilterID)
	filterOutput.Links.Dimensions.HRef = ""
	filterOutput.Links.FilterBlueprint.HRef = fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID)
	filterOutput.Links.FilterBlueprint.ID = filterBlueprintID
	filterOutput.LastUpdated = time.Now()

	// Clear out any event information to output document
	filterOutput.Events = models.Events{}

	// Downloads object should exist for filter output resource
	// even if it they are empty
	filterOutput.Downloads = &models.Downloads{
		CSV: &models.DownloadItem{},
		XLS: &models.DownloadItem{},
	}

	// Remove dimension url from output filter resource
	for i := range newFilter.Dimensions {
		filterOutput.Dimensions[i].URL = ""
	}

	if newFilter.Published == &models.Published {
		filterOutput.Published = &models.Published
	}

	if err := api.dataStore.CreateFilterOutput(&filterOutput); err != nil {
		log.ErrorC("unable to create filter output", err, log.Data{"filter_output": filterOutput})
		return models.Filter{}, err
	}

	return filterOutput, api.outputQueue.Queue(&filterOutput)
}

func (api *FilterAPI) getFilterOutputPreview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_output_id"]
	requestedLimit := r.URL.Query().Get("limit")

	logData := log.Data{"filter_output_id": filterID}
	log.Info("get filter output preview", logData)

	var limit = 20
	var err error

	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			logData["requested_limit"] = requestedLimit
			log.ErrorC("requested limit is not a number", err, logData)
			http.Error(w, errRequestLimitNotNumber.Error(), http.StatusBadRequest)
			return
		}
	}
	logData["limit"] = limit

	filterOutput, err := api.getOutput(r, filterID)
	if err != nil {
		log.ErrorC("failed to find filter output", err, logData)
		setErrorCode(w, err)
		return
	}

	logData["filter_output_dimensions"] = filterOutput.Dimensions

	if len(filterOutput.Dimensions) == 0 {
		log.ErrorC("no dimensions are present in the filter", errMissingDimensions, logData)
		http.Error(w, "no dimensions are present in the filter", http.StatusBadRequest)
		return
	}

	data, err := api.preview.GetPreview(filterOutput, limit)
	if err != nil {
		log.ErrorC("failed to query the graph database", err, logData)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.ErrorC("failed to marshal preview of filter ouput into bytes", err, logData)
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

	log.Info("preview filter output", logData)
}

func (api *FilterAPI) getOutput(r *http.Request, filterID string) (*models.Filter, error) {

	ctx := r.Context()
	logData := log.Data{"filter_output_id": filterID}

	output, err := api.dataStore.GetFilterOutput(filterID)
	if err != nil {
		log.Error(err, logData)
		return nil, err
	}

	logData["filter_blueprint_id"] = output.Links.FilterBlueprint.ID

	// Hide private download links if request is not authenticated
	if r.Header.Get(identity.DownloadServiceHeaderKey) != api.downloadServiceToken {

		log.Info("a valid download service token has not been provided. hiding private links", logData)

		if output.Downloads != nil {
			if output.Downloads.CSV != nil {
				output.Downloads.CSV.Private = ""
			}
			if output.Downloads.XLS != nil {
				output.Downloads.XLS.Private = ""
			}
		}
	} else {
		log.Info("a valid download service token has been provided. not hiding private links", logData)
	}

	//only return the filter if it is for published data or via authenticated request
	if output.Published != nil && *output.Published == models.Published || identity.IsCallerPresent(ctx) {
		return output, nil
	}

	log.Info("unauthenticated request to access unpublished filter output", logData)

	filter, err := api.getFilterBlueprint(ctx, output.Links.FilterBlueprint.ID)
	if err != nil {
		log.Error(errors.New("failed to retrieve filter blueprint"), logData)
		return nil, filters.ErrFilterOutputNotFound
	}

	//filter has been published since output was last requested, so update output and return
	if filter.Published != nil && *filter.Published == models.Published {
		output.Published = &models.Published
		if err := api.dataStore.UpdateFilterOutput(output); err != nil {
			log.Error(err, logData)
			return nil, filters.ErrFilterOutputNotFound
		}

		return output, nil
	}

	return nil, filters.ErrFilterOutputNotFound
}

func buildDownloadsObject(previousFilterOutput, filterOutput *models.Filter, downloadServiceURL string) *models.Filter {

	if filterOutput.Downloads == nil {
		filterOutput.Downloads = previousFilterOutput.Downloads
		return filterOutput
	}

	if filterOutput.Downloads.CSV != nil {

		filterOutput.Downloads.CSV.HRef = downloadServiceURL + "/downloads/filter-outputs/" + previousFilterOutput.FilterID + ".csv"

		if previousFilterOutput.Downloads != nil && previousFilterOutput.Downloads.CSV != nil {

			if filterOutput.Downloads.CSV.Size == "" {
				filterOutput.Downloads.CSV.Size = previousFilterOutput.Downloads.CSV.Size
			}
			if filterOutput.Downloads.CSV.Private == "" {
				filterOutput.Downloads.CSV.Private = previousFilterOutput.Downloads.CSV.Private
			}
			if filterOutput.Downloads.CSV.Public == "" {
				filterOutput.Downloads.CSV.Public = previousFilterOutput.Downloads.CSV.Public
			}
		}
	} else {
		if previousFilterOutput.Downloads != nil {
			filterOutput.Downloads.CSV = previousFilterOutput.Downloads.CSV
		}
	}

	if filterOutput.Downloads.XLS != nil {

		filterOutput.Downloads.XLS.HRef = downloadServiceURL + "/downloads/filter-outputs/" + previousFilterOutput.FilterID + ".xlsx"

		if previousFilterOutput.Downloads != nil && previousFilterOutput.Downloads.XLS != nil {

			if filterOutput.Downloads.XLS.Size == "" {
				filterOutput.Downloads.XLS.Size = previousFilterOutput.Downloads.XLS.Size
			}
			if filterOutput.Downloads.XLS.Private == "" {
				filterOutput.Downloads.XLS.Private = previousFilterOutput.Downloads.XLS.Private
			}
			if filterOutput.Downloads.XLS.Public == "" {
				filterOutput.Downloads.XLS.Public = previousFilterOutput.Downloads.XLS.Public
			}
		}
	} else {
		if previousFilterOutput.Downloads != nil {
			filterOutput.Downloads.XLS = previousFilterOutput.Downloads.XLS
		}
	}

	return filterOutput
}
