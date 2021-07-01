package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	errRequestLimitNotNumber = errors.New("requested limit is not a number")
	errMissingDimensions     = filters.NewBadRequestErr("no dimensions are present in the filter")
)

func (api *FilterAPI) getFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]
	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()

	log.Event(ctx, "getting filter output", log.INFO, logData)

	hideS3Links := r.Header.Get(dprequest.DownloadServiceHeaderKey) != api.downloadServiceToken
	filterOutput, err := api.getOutput(ctx, filterOutputID, hideS3Links)
	if err != nil {
		log.Event(ctx, "unable to get filter output", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
	logData["filter_output"] = filterOutput

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.Event(ctx, "failed to marshal filter output into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "got filter output", log.INFO, logData)
}

func (api *FilterAPI) updateFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()
	log.Event(ctx, "handling update filter output request", log.INFO, logData)

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Event(ctx, "unable to unmarshal request body", log.ERROR, log.Error(err), logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}
	logData["filter_output"] = filterOutput

	err = api.updateFilterOutput(ctx, filterOutputID, filterOutput)
	if err != nil {
		log.Event(ctx, "failed to update filter output", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
}

func (api *FilterAPI) updateFilterOutput(ctx context.Context, filterOutputID string, filterOutput *models.Filter) error {
	logData := log.Data{"filter_output_id": filterOutputID}

	log.Event(ctx, "updating filter output", log.INFO, logData)

	if !dprequest.IsCallerPresent(ctx) {
		log.Event(ctx, "failed to update filter output", log.ERROR, log.Error(filters.ErrUnauthorised), logData)
		return filters.ErrUnauthorised
	}

	// check filter output resource for current downloads and published flag
	previousFilterOutput, err := api.dataStore.GetFilterOutput(ctx, filterOutputID)
	if err != nil {
		log.Event(ctx, "unable to get current filter output", log.ERROR, log.Error(err), logData)
		return err
	}

	timestamp := previousFilterOutput.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	if err = filterOutput.ValidateFilterOutputUpdate(previousFilterOutput); err != nil {
		log.Event(ctx, "filter output failed validation", log.ERROR, log.Error(err), logData)
		return filters.NewForbiddenErr(err.Error())
	}

	filterOutput.FilterID = filterOutputID

	// Set the published flag to the value currently stored on filter output resources
	// unless the request contains an update to the flag
	if previousFilterOutput.Published != nil && *previousFilterOutput.Published == models.Published {
		filterOutput.Published = &models.Published
	}

	buildDownloadsObject(previousFilterOutput, filterOutput, api.downloadServiceURL)

	filterOutput.State = previousFilterOutput.State

	isNowStatusCompleted := false
	if downloadsAreGenerated(filterOutput) {
		log.Event(ctx, "downloads have been generated, setting filter output status to completed", log.INFO, logData)
		filterOutput.State = models.CompletedState
		isNowStatusCompleted = true
	}

	if err = api.dataStore.UpdateFilterOutput(ctx, filterOutput, timestamp); err != nil {
		log.Event(ctx, "unable to update filter output", log.ERROR, log.Error(err), logData)
		return err
	}

	// save the completed event after saving the filter output if its now complete
	if isNowStatusCompleted {
		log.Event(ctx, "filter output status is now completed, creating completed event", log.INFO, logData)

		completedEvent := &models.Event{
			Type: eventFilterOutputCompleted,
			Time: time.Now(),
		}

		if err = api.dataStore.AddEventToFilterOutput(ctx, filterOutput.FilterID, completedEvent); err != nil {
			log.Event(ctx, "failed to add event to filter output", log.ERROR, log.Error(err), logData)
			return err
		}
	}

	return nil
}

func downloadsAreGenerated(filterOutput *models.Filter) bool {
	if filterOutput.State == models.CompletedState {
		return true
	}

	// if all downloads are complete then set the filter state to complete
	if filterOutput.Downloads != nil &&
		filterOutput.Downloads.CSV != nil &&
		(filterOutput.Downloads.CSV.HRef != "" ||
			filterOutput.Downloads.CSV.Skipped) &&
		filterOutput.Downloads.XLS != nil &&
		(filterOutput.Downloads.XLS.HRef != "" ||
			filterOutput.Downloads.XLS.Skipped) {
		return true
	}

	return false
}

func (api *FilterAPI) getFilterOutputPreviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]
	requestedLimit := r.URL.Query().Get("limit")
	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()

	var limit = 20 // default if no limit is given
	var err error
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			logData["requested_limit"] = requestedLimit
			log.Event(ctx, "requested limit is not a number", log.ERROR, log.Error(err), logData)
			http.Error(w, errRequestLimitNotNumber.Error(), http.StatusBadRequest)
			return
		}
	}

	preview, err := api.getFilterOutputPreview(ctx, filterOutputID, limit)
	if err != nil {
		log.Event(ctx, "failed to get filter output preview", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(preview)
	if err != nil {
		log.Event(ctx, "failed to marshal preview of filter ouput into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "preview filter output", log.INFO, logData)
}

func (api *FilterAPI) getFilterOutputPreview(ctx context.Context, filterOutputID string, limit int) (*models.FilterPreview, error) {
	logData := log.Data{
		"filter_output_id": filterOutputID,
		"limit":            limit,
	}

	log.Event(ctx, "get filter output preview", log.INFO, logData)

	hideS3Links := true // do not require s3 links for preview

	filterOutput, err := api.getOutput(ctx, filterOutputID, hideS3Links)
	if err != nil {
		log.Event(ctx, "failed to find filter output", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	logData["filter_output_dimensions"] = filterOutput.Dimensions

	if len(filterOutput.Dimensions) == 0 {
		log.Event(ctx, "no dimensions are present in the filter", log.ERROR, log.Error(errMissingDimensions), logData)
		return nil, errMissingDimensions
	}

	filterOutputPreview, err := api.preview.GetPreview(ctx, filterOutput, limit)
	if err != nil {
		log.Event(ctx, "failed to query the graph database", log.ERROR, log.Error(err), logData)
		return nil, filters.ErrInternalError
	}

	return filterOutputPreview, nil
}

func (api *FilterAPI) getOutput(ctx context.Context, filterID string, hideS3Links bool) (*models.Filter, error) {
	logData := log.Data{"filter_output_id": filterID}

	output, err := api.dataStore.GetFilterOutput(ctx, filterID)
	if err != nil {
		log.Event(ctx, "error getting filter output", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	output.ID = output.FilterID
	var blueprintID string

	if output.Links.FilterBlueprint != nil {
		blueprintID = output.Links.FilterBlueprint.ID
	}

	logData["filter_blueprint_id"] = blueprintID

	// Hide private download links if request is not authenticated
	if hideS3Links {
		log.Event(ctx, "a valid download service token has not been provided. hiding links", log.INFO, logData)

		if output.Downloads != nil {
			if output.Downloads.CSV != nil {
				output.Downloads.CSV.Private = ""
				output.Downloads.CSV.Public = ""
			}
			if output.Downloads.XLS != nil {
				output.Downloads.XLS.Private = ""
				output.Downloads.XLS.Public = ""
			}
		}
	} else {
		log.Event(ctx, "a valid download service token has been provided. not hiding private links", log.INFO, logData)
	}

	// only return the filter if it is for published data or via authenticated request
	if output.Published != nil && *output.Published == models.Published || dprequest.IsCallerPresent(ctx) {
		return output, nil
	}

	log.Event(ctx, "unauthenticated request to access unpublished filter output", log.INFO, logData)

	filter, err := api.getFilterBlueprint(ctx, output.Links.FilterBlueprint.ID, mongo.AnyETag)
	if err != nil {
		log.Event(ctx, "failed to retrieve filter blueprint", log.ERROR, log.Error(err), logData)
		return nil, filters.ErrFilterOutputNotFound
	}

	// filter has been published since output was last requested, so update output and return
	if filter.Published != nil && *filter.Published == models.Published {
		output.Published = &models.Published
		if err := api.dataStore.UpdateFilterOutput(ctx, output, output.UniqueTimestamp); err != nil {
			log.Event(ctx, "error updating filter output", log.ERROR, log.Error(err), logData)
			return nil, filters.ErrFilterOutputNotFound
		}

		return output, nil
	}

	return nil, filters.ErrFilterOutputNotFound
}

func buildDownloadsObject(previousFilterOutput, filterOutput *models.Filter, downloadServiceURL string) {
	if filterOutput.Downloads == nil {
		filterOutput.Downloads = previousFilterOutput.Downloads
		return
	}

	if previousFilterOutput.Downloads != nil && previousFilterOutput.Downloads.CSV != nil {
		filterOutput.Downloads.CSV = buildDownloadItem(filterOutput.Downloads.CSV, previousFilterOutput.Downloads.CSV)

	}

	if previousFilterOutput.Downloads != nil && previousFilterOutput.Downloads.XLS != nil {
		filterOutput.Downloads.XLS = buildDownloadItem(filterOutput.Downloads.XLS, previousFilterOutput.Downloads.XLS)

	}

	baseHref := downloadServiceURL + "/downloads/filter-outputs/" + previousFilterOutput.FilterID
	if filterOutput.Downloads.CSV != nil && !filterOutput.Downloads.CSV.Skipped && len(filterOutput.Downloads.CSV.Size) > 0 {
		filterOutput.Downloads.CSV.HRef = baseHref + ".csv"

	}
	if filterOutput.Downloads.XLS != nil && !filterOutput.Downloads.XLS.Skipped && len(filterOutput.Downloads.XLS.Size) > 0 {
		filterOutput.Downloads.XLS.HRef = baseHref + ".xlsx"
	}
}

func buildDownloadItem(new, old *models.DownloadItem) *models.DownloadItem {
	if new == nil {
		return old
	}

	if new.Skipped {
		return &models.DownloadItem{
			Skipped: true,
		}
	}

	if new.Size == "" {
		new.Size = old.Size
	}
	if new.Private == "" {
		new.Private = old.Private
	}
	if new.Public == "" {
		new.Public = old.Public
	}

	return new
}

func (api *FilterAPI) addEventHandler(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()
	log.Event(ctx, "add event to filter output endpoint called", log.INFO, logData)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Event(ctx, "failed to read request body", log.ERROR, log.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := &models.Event{}
	err = json.Unmarshal([]byte(bytes), event)
	if err != nil {
		log.Event(ctx, "failed to parse json body", log.ERROR, log.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logData["event"] = event
	log.Event(ctx, "adding event to filter output", log.INFO, logData)

	err = api.addEvent(ctx, filterOutputID, event)
	if err != nil {
		log.Event(ctx, "failed to add event to filter output", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Event(ctx, "added event to filter output", log.INFO, logData)
}

func (api *FilterAPI) addEvent(ctx context.Context, filterOutputID string, event *models.Event) error {
	if event.Type == "" {
		return filters.NewBadRequestErr("event type cannot be empty")
	}

	return api.dataStore.AddEventToFilterOutput(ctx, filterOutputID, event)
}
