package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-net/v2/links"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

//nolint:gocyclo // high cyclomactic complexity not in scope for maintenance
func (api *FilterAPI) getFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]
	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()
	log.Info(ctx, "getting filter output", logData)

	hideS3Links := r.Header.Get(dprequest.DownloadServiceHeaderKey) != api.downloadServiceToken
	filterOutput, err := api.getOutput(ctx, filterOutputID, hideS3Links)
	if err != nil {
		log.Error(ctx, "unable to get filter output", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["filter_output"] = filterOutput

	if api.enableURLRewriting {
		filterAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)
		datasetAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.DatasetAPIURL)
		downloadLinksBuilder := links.FromHeadersOrDefaultDownload(&r.Header, api.downloadServiceURL, api.ExternalDownloadServiceURL)

		if filterOutput.Links.Self != nil && filterOutput.Links.Self.HRef != "" {
			newLink, err := filterAPILinksBuilder.BuildLink(filterOutput.Links.Self.HRef)
			if err != nil {
				log.Error(ctx, "failed to rewrite filter output self link", err, logData,
					log.Data{"link_type": "Self", "original_link": filterOutput.Links.Self.HRef})
				setErrorCode(w, err)
				return
			}
			filterOutput.Links.Self.HRef = newLink
		}

		if filterOutput.Links.FilterBlueprint != nil && filterOutput.Links.FilterBlueprint.HRef != "" {
			newLink, err := filterAPILinksBuilder.BuildLink(filterOutput.Links.FilterBlueprint.HRef)
			if err != nil {
				log.Error(ctx, "failed to rewrite filter output filterBlueprint link", err, logData,
					log.Data{"link_type": "FilterBlueprint", "original_link": filterOutput.Links.FilterBlueprint.HRef})
				setErrorCode(w, err)
				return
			}
			filterOutput.Links.FilterBlueprint.HRef = newLink
		}

		if filterOutput.Links.Version != nil && filterOutput.Links.Version.HRef != "" {
			newLink, err := datasetAPILinksBuilder.BuildLink(filterOutput.Links.Version.HRef)
			if err != nil {
				log.Error(ctx, "failed to rewrite filter output version link", err, logData,
					log.Data{"link_type": "Version", "original_link": filterOutput.Links.Version.HRef})
				setErrorCode(w, err)
				return
			}
			filterOutput.Links.Version.HRef = newLink
		}

		if filterOutput.Downloads != nil {
			if filterOutput.Downloads.CSV != nil && filterOutput.Downloads.CSV.HRef != "" {
				newDownloadLink, err := downloadLinksBuilder.BuildDownloadLink(filterOutput.Downloads.CSV.HRef)
				if err != nil {
					log.Error(ctx, "failed to rewrite CSV download link", err, logData,
						log.Data{"link_type": "CSV", "original_link": filterOutput.Downloads.CSV.HRef})
					setErrorCode(w, err)
					return
				}
				filterOutput.Downloads.CSV.HRef = newDownloadLink
			}
			if filterOutput.Downloads.XLS != nil && filterOutput.Downloads.XLS.HRef != "" {
				newDownloadLink, err := downloadLinksBuilder.BuildDownloadLink(filterOutput.Downloads.XLS.HRef)
				if err != nil {
					log.Error(ctx, "failed to rewrite XLS download link", err, logData,
						log.Data{"link_type": "XLS", "original_link": filterOutput.Downloads.XLS.HRef})
					setErrorCode(w, err)
					return
				}
				filterOutput.Downloads.XLS.HRef = newDownloadLink
			}
		}
	}

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.Error(ctx, "failed to marshal filter output into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "got filter output", logData)
}

func (api *FilterAPI) updateFilterOutputHandler(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	ctx := r.Context()
	log.Info(ctx, "handling update filter output request", logData)

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(ctx, "unable to unmarshal request body", err, logData)
		http.Error(w, BadRequest, http.StatusBadRequest)
		return
	}
	logData["filter_output"] = filterOutput

	err = api.updateFilterOutput(ctx, filterOutputID, filterOutput)
	if err != nil {
		log.Error(ctx, "failed to update filter output", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
}

func (api *FilterAPI) updateFilterOutput(ctx context.Context, filterOutputID string, filterOutput *models.Filter) error {
	logData := log.Data{"filter_output_id": filterOutputID}

	log.Info(ctx, "updating filter output", logData)

	if !dprequest.IsCallerPresent(ctx) {
		log.Error(ctx, "failed to update filter output", filters.ErrUnauthorised, logData)
		return filters.ErrUnauthorised
	}

	// check filter output resource for current downloads and published flag
	previousFilterOutput, err := api.dataStore.GetFilterOutput(ctx, filterOutputID)
	if err != nil {
		log.Error(ctx, "unable to get current filter output", err, logData)
		return err
	}

	timestamp := previousFilterOutput.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	if err = filterOutput.ValidateFilterOutputUpdate(previousFilterOutput); err != nil {
		log.Error(ctx, "filter output failed validation", err, logData)
		return filters.NewForbiddenErr(err.Error())
	}

	filterOutput.FilterID = filterOutputID

	// Set the published flag to the value currently stored on filter output resources
	// unless the request contains an update to the flag
	if previousFilterOutput.Published != nil && *previousFilterOutput.Published == models.Published {
		filterOutput.Published = &models.Published
	}

	BuildDownloadsObject(previousFilterOutput, filterOutput, api.downloadServiceURL.String())

	filterOutput.State = previousFilterOutput.State

	isNowStatusCompleted := false
	if downloadsAreGenerated(filterOutput) {
		log.Info(ctx, "downloads have been generated, setting filter output status to completed", logData)
		filterOutput.State = models.CompletedState
		isNowStatusCompleted = true
	}

	if err = api.dataStore.UpdateFilterOutput(ctx, filterOutput, timestamp); err != nil {
		log.Error(ctx, "unable to update filter output", err, logData)
		return err
	}

	filterOutput.RemoveDuplicateEvents(previousFilterOutput)
	if err = api.addEvents(ctx, filterOutput.Events, filterOutputID, isNowStatusCompleted); err != nil {
		log.Error(ctx, "unable to add events to filter output", err, logData)
		return err
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

func (api *FilterAPI) getOutput(ctx context.Context, filterID string, hideS3Links bool) (*models.Filter, error) {
	logData := log.Data{"filter_output_id": filterID}

	output, err := api.dataStore.GetFilterOutput(ctx, filterID)
	if err != nil {
		log.Error(ctx, "error getting filter output", err, logData)
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
		log.Info(ctx, "a valid download service token has not been provided. hiding links", logData)

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
		log.Info(ctx, "a valid download service token has been provided. not hiding private links", logData)
	}

	// only return the filter if it is for published data or via authenticated request
	if output.Published != nil && *output.Published == models.Published || dprequest.IsCallerPresent(ctx) {
		return output, nil
	}

	log.Info(ctx, "unauthenticated request to access unpublished filter output", logData)

	filter, err := api.getFilterBlueprint(ctx, output.Links.FilterBlueprint.ID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "failed to retrieve filter blueprint", err, logData)
		return nil, filters.ErrFilterOutputNotFound
	}

	// filter has been published since output was last requested, so update output and return
	if filter.Published != nil && *filter.Published == models.Published {
		output.Published = &models.Published
		if err := api.dataStore.UpdateFilterOutput(ctx, output, output.UniqueTimestamp); err != nil {
			log.Error(ctx, "error updating filter output", err, logData)
			return nil, filters.ErrFilterOutputNotFound
		}

		return output, nil
	}

	return nil, filters.ErrFilterOutputNotFound
}

func BuildDownloadsObject(previousFilterOutput, filterOutput *models.Filter, downloadServiceURL string) {
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
	if filterOutput.Downloads.CSV != nil && !filterOutput.Downloads.CSV.Skipped && filterOutput.Downloads.CSV.Size != "" {
		filterOutput.Downloads.CSV.HRef = baseHref + ".csv"
	}

	if filterOutput.Downloads.XLS != nil && !filterOutput.Downloads.XLS.Skipped && filterOutput.Downloads.XLS.Size != "" {
		filterOutput.Downloads.XLS.HRef = baseHref + ".xlsx"
	}
}

//nolint:gocritic // shadowing of predeclared identifier: new is acceptable here
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
	log.Info(ctx, "add event to filter output endpoint called", logData)

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(ctx, "failed to read request body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := &models.Event{}
	err = json.Unmarshal(bytes, event)
	if err != nil {
		log.Error(ctx, "failed to parse json body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logData["event"] = event
	log.Info(ctx, "adding event to filter output", logData)

	err = api.addEvents(ctx, []*models.Event{event}, filterOutputID, false)
	if err != nil {
		log.Error(ctx, "failed to add event to filter output", err, logData)
		setErrorCode(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Info(ctx, "added event to filter output", logData)
}

func (api *FilterAPI) addEvents(ctx context.Context, events []*models.Event, id string, completed bool) error {
	for _, e := range events {
		if err := e.Validate(); err != nil {
			return filters.NewBadRequestErr(err.Error())
		}

		if err := api.dataStore.AddEventToFilterOutput(ctx, id, e); err != nil {
			return err
		}
	}

	// save the completed event after saving the filter output if its now complete
	if completed {
		completedEvent := &models.Event{
			Type: models.EventFilterOutputCompleted,
			Time: time.Now(),
		}

		if err := api.dataStore.AddEventToFilterOutput(ctx, id, completedEvent); err != nil {
			return err
		}
	}

	return nil
}
