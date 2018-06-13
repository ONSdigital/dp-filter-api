package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"strconv"

	"context"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/preview"
	"github.com/ONSdigital/go-ns/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

var (
	errRequestLimitNotNumber = errors.New("requested limit is not a number")
	errMissingDimensions     = filters.NewBadRequestErr("no dimensions are present in the filter")
)

const (
	// audit actions
	getFilterOutputAction    = "getFilterOutput"
	updateFilterOutputAction = "updateFilterOutput"
	getFilterPreviewAction   = "getFilterPreview"
)

func (api *FilterAPI) getFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.InfoCtx(r.Context(), "getting filter output", logData)

	auditParams := common.Params{"filter_output_id": filterOutputID}
	if auditErr := api.auditor.Record(r.Context(), getFilterOutputAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getFilterOutputAction, actionAttempted, w, auditErr, logData)
		return
	}

	hideS3Links := r.Header.Get(common.DownloadServiceHeaderKey) != api.downloadServiceToken
	filterOutput, err := api.getOutput(r.Context(), filterOutputID, hideS3Links)
	if err != nil {
		log.ErrorC("unable to get filter output", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterOutputAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getFilterOutputAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}
	logData["filter_output"] = filterOutput

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.ErrorC("failed to marshal filter output into bytes", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterOutputAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getFilterOutputAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getFilterOutputAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getFilterOutputAction, actionSuccessful, w, auditErr, logData)
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

	log.InfoCtx(r.Context(), "got filter output", logData)
}

func (api *FilterAPI) updateFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.InfoCtx(r.Context(), "updating filter output", logData)

	auditParams := common.Params{"filter_output_id": filterOutputID}
	if auditErr := api.auditor.Record(r.Context(), updateFilterOutputAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), updateFilterOutputAction, actionAttempted, w, auditErr, logData)
		return
	}

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		if auditErr := api.auditor.Record(r.Context(), updateFilterOutputAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), updateFilterOutputAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}
	logData["filter_output"] = filterOutput

	err = api.updateFilterOutput(r.Context(), filterOutputID, filterOutput)
	if err != nil {
		log.ErrorC("failed to update filter output", err, logData)
		if auditErr := api.auditor.Record(r.Context(), updateFilterOutputAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), updateFilterOutputAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), updateFilterOutputAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(r.Context(), updateFilterOutputAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
}

func (api *FilterAPI) updateFilterOutput(ctx context.Context, filterOutputID string, filterOutput *models.Filter) error {

	logData := log.Data{"filter_output_id": filterOutputID}
	log.InfoCtx(ctx, "updating filter output", logData)

	if !common.IsCallerPresent(ctx) {
		log.ErrorC("failed to update filter output", filters.ErrUnauthorised, logData)
		return filters.ErrUnauthorised
	}

	// check filter output resource for current downloads and published flag
	previousFilterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.ErrorC("unable to get current filter output", err, logData)
		return err
	}

	timestamp := previousFilterOutput.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	if err = filterOutput.ValidateFilterOutputUpdate(previousFilterOutput); err != nil {
		log.ErrorC("filter output failed validation", err, logData)
		return filters.NewForbiddenErr(err.Error())
	}

	filterOutput.FilterID = filterOutputID

	// Set the published flag to the value currently stored on filter output resources
	// unless the request contains an update to the flag
	if previousFilterOutput.Published != nil && *previousFilterOutput.Published == models.Published {
		filterOutput.Published = &models.Published
	}

	buildDownloadsObject(previousFilterOutput, filterOutput, api.downloadServiceURL)

	isNowStatusCompleted := false
	if downloadsAreGenerated(filterOutput) {
		log.InfoCtx(ctx, "downloads have been generated, setting filter output status to completed", logData)
		filterOutput.State = models.CompletedState
		isNowStatusCompleted = true
	}

	if err = api.dataStore.UpdateFilterOutput(filterOutput, timestamp); err != nil {
		log.ErrorC("unable to update filter output", err, logData)
		return err
	}

	// save the completed event after saving the filter output if its now complete
	if isNowStatusCompleted {
		log.InfoCtx(ctx, "filter output status is now completed, creating completed event", logData)

		completedEvent := &models.Event{
			Type: eventFilterOutputCompleted,
			Time: time.Now(),
		}

		if err = api.dataStore.AddEventToFilterOutput(filterOutput.FilterID, completedEvent); err != nil {
			log.ErrorCtx(ctx, errors.Wrap(err, "failed to add event to filter output"), logData)
			return err
		}
	}

	return nil
}

func downloadsAreGenerated(filterOutput *models.Filter) bool {
	if filterOutput.State != models.CompletedState {

		// if all downloads are complete then set the filter state to complete
		if filterOutput.Downloads != nil &&
			filterOutput.Downloads.CSV != nil &&
			filterOutput.Downloads.CSV.HRef != "" &&
			filterOutput.Downloads.XLS != nil &&
			filterOutput.Downloads.XLS.HRef != "" {
			return true
		}
	}

	return false
}

func (api *FilterAPI) getFilterOutputPreviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]
	requestedLimit := r.URL.Query().Get("limit")

	logData := log.Data{"filter_output_id": filterOutputID}

	auditParams := common.Params{"filter_output_id": filterOutputID}
	if auditErr := api.auditor.Record(r.Context(), getFilterPreviewAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getFilterPreviewAction, actionAttempted, w, auditErr, logData)
		return
	}

	var limit = 20 // default if no limit is given
	var err error
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			logData["requested_limit"] = requestedLimit
			log.ErrorC("requested limit is not a number", err, logData)
			if auditErr := api.auditor.Record(r.Context(), getFilterPreviewAction, actionUnsuccessful, auditParams); auditErr != nil {
				handleAuditingFailure(r.Context(), getFilterPreviewAction, actionUnsuccessful, w, auditErr, logData)
				return
			}
			http.Error(w, errRequestLimitNotNumber.Error(), http.StatusBadRequest)
			return
		}
	}

	preview, err := api.getFilterOutputPreview(r.Context(), filterOutputID, limit)
	if err != nil {
		log.ErrorC("failed to get filter output preview", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterPreviewAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getFilterPreviewAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(preview)
	if err != nil {
		log.ErrorC("failed to marshal preview of filter ouput into bytes", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterPreviewAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(r.Context(), getFilterPreviewAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(r.Context(), getFilterPreviewAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(r.Context(), getFilterPreviewAction, actionSuccessful, w, auditErr, logData)
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

	log.InfoCtx(r.Context(), "preview filter output", logData)
}

func (api *FilterAPI) getFilterOutputPreview(ctx context.Context, filterOutputID string, limit int) (*preview.FilterPreview, error) {

	logData := log.Data{
		"filter_output_id": filterOutputID,
		"limit":            limit,
	}
	log.InfoCtx(ctx, "get filter output preview", logData)

	hideS3Links := true // do not require s3 links for preview
	filterOutput, err := api.getOutput(ctx, filterOutputID, hideS3Links)
	if err != nil {
		log.ErrorC("failed to find filter output", err, logData)
		return nil, err
	}

	logData["filter_output_dimensions"] = filterOutput.Dimensions

	if len(filterOutput.Dimensions) == 0 {
		log.ErrorC(errMissingDimensions.Error(), errMissingDimensions, logData)
		return nil, errMissingDimensions
	}

	filterOutputPreview, err := api.preview.GetPreview(filterOutput, limit)
	if err != nil {
		log.ErrorC("failed to query the graph database", err, logData)
		return nil, filters.ErrInternalError
	}

	return filterOutputPreview, nil
}

func (api *FilterAPI) getOutput(ctx context.Context, filterID string, hideS3Links bool) (*models.Filter, error) {

	logData := log.Data{"filter_output_id": filterID}

	output, err := api.dataStore.GetFilterOutput(filterID)
	if err != nil {
		log.ErrorCtx(ctx, err, logData)
		return nil, err
	}

	logData["filter_blueprint_id"] = output.Links.FilterBlueprint.ID

	// Hide private download links if request is not authenticated
	if hideS3Links {

		log.InfoCtx(ctx, "a valid download service token has not been provided. hiding links", logData)

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
		log.InfoCtx(ctx, "a valid download service token has been provided. not hiding private links", logData)
	}

	//only return the filter if it is for published data or via authenticated request
	if output.Published != nil && *output.Published == models.Published || common.IsCallerPresent(ctx) {
		return output, nil
	}

	log.InfoCtx(ctx, "unauthenticated request to access unpublished filter output", logData)

	filter, err := api.getFilterBlueprint(ctx, output.Links.FilterBlueprint.ID)
	if err != nil {
		log.ErrorCtx(ctx, errors.New("failed to retrieve filter blueprint"), logData)
		return nil, filters.ErrFilterOutputNotFound
	}

	//filter has been published since output was last requested, so update output and return
	if filter.Published != nil && *filter.Published == models.Published {
		output.Published = &models.Published
		if err := api.dataStore.UpdateFilterOutput(output, output.UniqueTimestamp); err != nil {
			log.ErrorCtx(ctx, err, logData)
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
}

func (api *FilterAPI) addEventHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.InfoCtx(r.Context(), "add event to filter output endpoint called", logData)

	defer r.Body.Close()

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "failed to read request body"), nil)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := &models.Event{}
	err = json.Unmarshal([]byte(bytes), event)
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "failed to parse json body"), nil)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logData["event"] = event
	log.InfoCtx(r.Context(), "adding event to filter output", logData)

	err = api.addEvent(filterOutputID, event)
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "failed to add event to filter output"), logData)
		setErrorCode(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.InfoCtx(r.Context(), "added event to filter output", logData)
}

func (api *FilterAPI) addEvent(filterOutputID string, event *models.Event) error {

	if event.Type == "" {
		return filters.NewBadRequestErr("event type cannot be empty")
	}

	return api.dataStore.AddEventToFilterOutput(filterOutputID, event)
}
