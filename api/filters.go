package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	datasetAPI "github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

var (
	internalError = "Failed to process the request due to an internal error"
	badRequest    = "Bad request - Invalid request body"

	statusBadRequest          = "bad request"
	statusUnprocessableEntity = "unprocessable entity"

	incorrectDimensionOptions = regexp.MustCompile("incorrect dimension options chosen")
	incorrectDimension        = regexp.MustCompile("incorrect dimensions chosen")

	publishedState = "published"
)

const (
	filterSubmitted = "true"

	eventFilterOutputCreated   = "FilterOutputCreated"
	eventFilterOutputCompleted = "FilterOutputCompleted"
)

func (api *FilterAPI) postFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	submitted := r.FormValue("submitted")
	logData := log.Data{"submitted": submitted}
	ctx := r.Context()
	log.Event(ctx, "create filter blueprint", log.INFO, logData)

	filter, err := models.CreateNewFilter(r.Body)
	if err != nil {
		log.Event(ctx, "unable to unmarshal request body", log.ERROR, log.Error(err), logData)
		if err, ok := err.(models.DuplicateDimensionError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, badRequest, http.StatusBadRequest)
		}
		return
	}

	newFilter, err := api.createFilterBlueprint(ctx, filter, submitted)
	if err != nil {
		log.Event(ctx, "failed to create new filter", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
	log.Event(ctx, "created filter blueprint", log.INFO, logData)

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newFilter.ETag)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
}

func (api *FilterAPI) createFilterBlueprint(ctx context.Context, filter *models.NewFilter, submitted string) (*models.Filter, error) {

	newFilter := &models.Filter{}
	logData := log.Data{}

	if err := filter.ValidateNewFilter(); err != nil {
		logData["filter_parameters"] = filter
		log.Event(ctx, "filter parameters failed validation", log.ERROR, log.Error(err), logData)
		return nil, filters.ErrBadRequest
	}

	// Create unique id
	u, err := uuid.NewV4()
	if err != nil {
		log.Event(ctx, "failed to create a new UUID for filter blueprint", log.ERROR, log.Error(err), logData)
		return nil, err
	}
	newFilter.FilterID = u.String()
	newFilter.Dimensions = filter.Dimensions
	logData["new_filter"] = newFilter

	// add version information from datasetAPI
	version, err := api.getVersion(ctx, filter.Dataset)
	if err != nil {
		log.Event(ctx, "unable to retrieve version document", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	if version.State != publishedState && !dprequest.IsCallerPresent(ctx) {
		log.Event(ctx, "unauthenticated request to filter unpublished version", log.INFO, log.Data{"dataset": *filter.Dataset, "state": version.State})
		return nil, filters.ErrVersionNotFound
	}

	if version.State == publishedState {
		newFilter.Published = &models.Published
	}

	links := models.LinkMap{
		Dimensions: &models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s/dimensions", api.host, newFilter.FilterID),
		},
		Self: &models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s", api.host, newFilter.FilterID),
		},
		Version: &models.LinkObject{
			HRef: version.Links.Self.URL,
			ID:   strconv.Itoa(version.Version),
		},
	}

	newFilter.Links = links
	newFilter.InstanceID = version.ID
	newFilter.Dataset = filter.Dataset
	logData["new_filter"] = newFilter

	if err = api.checkFilterOptions(ctx, newFilter, version); err != nil {
		log.Event(ctx, "failed to select valid filter options", log.ERROR, log.Error(err), logData)
		return nil, filters.NewBadRequestErr(err.Error())
	}

	newFilter, err = api.dataStore.AddFilter(newFilter)
	if err != nil {
		log.Event(ctx, "failed to create new filter blueprint", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	if submitted == filterSubmitted {
		var filterOutput models.Filter
		// Create filter output resource and use filter id, dataset, edition and version to pass into kafka
		filterOutput, err = api.createFilterOutputResource(ctx, newFilter, newFilter.FilterID)
		if err != nil {
			log.Event(ctx, "failed to create new filter output", log.ERROR, log.Error(err), logData)
			return nil, err
		}
		logData["filter_output_id"] = filterOutput.FilterID

		newFilter.Links = links
		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID

		logData["new_filter"] = newFilter
		log.Event(ctx, "filter output id sent in message to kafka", log.INFO, logData)
	}

	return newFilter, nil
}

func (api *FilterAPI) getFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	ctx := r.Context()
	log.Event(ctx, "getting filter blueprint", log.INFO, logData)

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterID, mongo.AnyETag)
	if err != nil {
		log.Event(ctx, "unable to get filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	filterBlueprint.Dimensions = nil
	logData["filter_blueprint"] = filterBlueprint

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filterBlueprint.ETag)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
	}

	log.Event(ctx, "got filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) putFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	submitted := r.URL.Query().Get("submitted")
	logData := log.Data{"filter_blueprint_id": filterID, "submitted": submitted}
	ctx := r.Context()
	log.Event(ctx, "updating filter blueprint", log.INFO, logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Event(ctx, "missing header", log.ERROR, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		// When filter blueprint has query parameter `submitted` set to true then
		// request can have an empty json in body for this PUT request
		if submitted != filterSubmitted || err != models.ErrorNoData {
			log.Event(ctx, "unable to unmarshal request body", log.ERROR, log.Error(err), logData)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
	}
	filter.FilterID = filterID

	newFilter, err := api.updateFilterBlueprint(ctx, filter, submitted, eTag)
	if err != nil {
		log.Event(ctx, "failed to update filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
	log.Event(ctx, "filter blueprint updated", log.INFO, logData)

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Event(ctx, "failed to marshal updated filter blueprint into bytes", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newFilter.ETag)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}
}

func (api *FilterAPI) updateFilterBlueprint(ctx context.Context, filter *models.Filter, submitted, eTag string) (*models.Filter, error) {

	logData := log.Data{"filter_blueprint_id": filter.FilterID, "submitted": submitted}
	log.Event(ctx, "updating filter blueprint", log.INFO, logData)
	logData["filter_update"] = filter

	if err := models.ValidateFilterBlueprintUpdate(filter); err != nil {
		log.Event(ctx, "filter blueprint failed validation", log.ERROR, log.Error(err), logData)
		return nil, filters.ErrBadRequest
	}

	currentFilter, err := api.getFilterBlueprint(ctx, filter.FilterID, eTag)
	if err != nil {
		log.Event(ctx, "unable to get filter blueprint", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	timestamp := currentFilter.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	logData["current_filter"] = currentFilter

	newFilter, versionHasChanged := createNewFilter(filter, currentFilter)
	logData["new_filter"] = newFilter

	if versionHasChanged {
		log.Event(ctx, "finding new version details for filter after version change", log.INFO, logData)

		version, err := api.getVersion(ctx, newFilter.Dataset)
		if err != nil {
			log.Event(ctx, "unable to retrieve version document", log.ERROR, log.Error(err), logData)
			return nil, filters.NewBadRequestErr(err.Error())
		}

		newFilter.Published = &models.Unpublished
		if version.State == "published" {
			newFilter.Published = &models.Published
		}

		newFilter.InstanceID = version.ID
		newFilter.Links.Version.HRef = version.Links.Self.URL

		// Check existing dimensions work for new version
		if err = api.checkFilterOptions(ctx, newFilter, version); err != nil {
			log.Event(ctx, "failed to select valid filter options", log.ERROR, log.Error(err), logData)
			return nil, filters.NewBadRequestErr(err.Error())
		}
	}

	newFilter.ETag, err = api.dataStore.UpdateFilter(newFilter, timestamp, eTag, currentFilter)
	if err != nil {
		log.Event(ctx, "unable to update filter blueprint", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	if submitted == filterSubmitted {
		outputFilter := newFilter

		var filterOutput models.Filter
		// Create filter output resource and use id to pass into kafka
		filterOutput, err = api.createFilterOutputResource(ctx, outputFilter, filter.FilterID)
		if err != nil {
			log.Event(ctx, "failed to create new filter output", log.ERROR, log.Error(err), logData)
			return nil, err
		}
		logData["filter_output_id"] = filterOutput.FilterID

		log.Event(ctx, "filter output id sent in message to kafka", log.INFO, logData)

		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID
	}

	return newFilter, nil
}

func (api *FilterAPI) getFilterBlueprint(ctx context.Context, filterID, eTag string) (*models.Filter, error) {

	logData := log.Data{"filter_blueprint_id": filterID}

	currentFilter, err := api.dataStore.GetFilter(filterID, mongo.AnyETag)
	if err != nil {
		log.Event(ctx, "error getting filter", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	if eTag != mongo.AnyETag && currentFilter.ETag != eTag {
		return nil, filters.ErrFilterBlueprintConflict
	}

	//only return the filter if it is for published data or via authenticated request
	if currentFilter.Published != nil && *currentFilter.Published == models.Published || dprequest.IsCallerPresent(ctx) {
		return currentFilter, nil
	}

	log.Event(ctx, "unauthenticated request to access unpublished filter", log.INFO, logData)

	version, err := api.getVersion(ctx, currentFilter.Dataset)
	if err != nil {
		log.Event(ctx, "failed to retrieve version from dataset api", log.ERROR, log.Error(err), logData)
		return nil, err
	}

	// version has been published since filter was last requested, so update filter and return
	if version.State == publishedState {
		filter := currentFilter
		filter.Published = &models.Published
		filter.ETag, err = api.dataStore.UpdateFilter(filter, filter.UniqueTimestamp, currentFilter.ETag, currentFilter)
		if err != nil {
			log.Event(ctx, "error updating filter", log.ERROR, log.Error(err), logData)
			if err == filters.ErrFilterBlueprintConflict {
				return nil, filters.ErrFilterBlueprintConflict
			}
			return nil, filters.ErrFilterBlueprintNotFound
		}

		return filter, nil
	}

	// not authenticated, so return not found
	return nil, filters.ErrFilterBlueprintNotFound
}

func (api *FilterAPI) checkFilterOptions(ctx context.Context, newFilter *models.Filter, version *datasetAPI.Version) error {
	logData := log.Data{"new_filter": newFilter, "version": version.Version}
	log.Event(ctx, "check filter dimension options before calling dataset api", log.INFO, logData)

	// Call dimensions list endpoint
	datasetDimensions, err := api.getDimensions(ctx, newFilter.Dataset)
	if err != nil {
		log.Event(ctx, "failed to retrieve a list of dimensions from the dataset API", log.ERROR, log.Error(err), logData)
		return err
	}

	logData["dataset_dimensions_total"] = len(datasetDimensions.Items)
	if len(datasetDimensions.Items) > 30 {
		logData["dataset_dimensions_first"] = datasetDimensions.Items[0]
	} else {
		logData["dataset_dimensions"] = datasetDimensions
	}

	log.Event(ctx, "dimensions retrieved from dataset API", log.INFO, logData)

	if err = models.ValidateFilterDimensions(newFilter.Dimensions, datasetDimensions); err != nil {
		log.Event(ctx, "filter dimensions failed validation", log.ERROR, log.Error(err), logData)
		return err
	}
	log.Event(ctx, "successfully validated filter dimensions", log.INFO, logData)

	// check options for all dimensions in the filter
	for _, filterDimension := range newFilter.Dimensions {
		if err := api.checkNewFilterDimensionOptions(ctx, filterDimension, newFilter.Dataset, logData); err != nil {
			return err
		}
	}
	return nil
}

func (api *FilterAPI) getVersion(ctx context.Context, dataset *models.Dataset) (*datasetAPI.Version, error) {

	dimensions, err := api.datasetAPI.GetVersion(ctx,
		"",
		api.serviceAuthToken,
		"",
		"",
		dataset.ID,
		dataset.Edition,
		strconv.Itoa(dataset.Version))

	if err != nil {
		if apiErr, ok := err.(*datasetAPI.ErrInvalidDatasetAPIResponse); ok {
			switch apiErr.Code() {
			case http.StatusNotFound:
				return nil, filters.ErrVersionNotFound
			}
		}

		return nil, err
	}

	return &dimensions, nil
}

func getUserAuthToken(ctx context.Context) string {

	if dprequest.IsFlorenceIdentityPresent(ctx) {
		return ctx.Value(dprequest.FlorenceIdentityKey).(string)
	}

	return ""
}

func getCollectionID(ctx context.Context) string {

	rawKeyValue := ctx.Value(dprequest.CollectionIDHeaderKey)

	if rawKeyValue != nil {
		collectionID := rawKeyValue.(string)
		return collectionID
	}

	return ""
}

func (api *FilterAPI) getDimensions(ctx context.Context, dataset *models.Dataset) (*datasetAPI.VersionDimensions, error) {

	dimensions, err := api.datasetAPI.GetVersionDimensions(ctx,
		getUserAuthToken(ctx),
		api.serviceAuthToken,
		getCollectionID(ctx),
		dataset.ID,
		dataset.Edition,
		strconv.Itoa(dataset.Version))
	if err != nil {
		if apiErr, ok := err.(*datasetAPI.ErrInvalidDatasetAPIResponse); ok {
			switch apiErr.Code() {
			case http.StatusNotFound:
				return nil, filters.ErrDimensionsNotFound
			}
		}

		return nil, err
	}

	return &dimensions, nil
}

// getDimensionOptionsBatchProcess calls dataset API GetOptionsBatchProcess with the provided batch processor
func (api *FilterAPI) getDimensionOptionsBatchProcess(ctx context.Context, dimension models.Dimension, dataset *models.Dataset, processBatch datasetAPI.OptionsBatchProcessor) error {

	// if no options are defined, there is nothing to validate against dataset API
	if dimension.Options == nil || len(dimension.Options) == 0 {
		return nil
	}

	// get encoded IDs so that they can be used as query paramters
	encodedIDs := dimension.EncodedOptions()

	// validate the options with Dataset API, in batches
	err := api.datasetAPI.GetOptionsBatchProcess(ctx,
		getUserAuthToken(ctx),
		api.serviceAuthToken,
		getCollectionID(ctx),
		dataset.ID,
		dataset.Edition,
		strconv.Itoa(dataset.Version),
		dimension.Name,
		&encodedIDs,
		processBatch,
		api.maxDatasetOptions,
		api.BatchMaxWorkers)

	if err != nil {
		if apiErr, ok := err.(*datasetAPI.ErrInvalidDatasetAPIResponse); ok {
			switch apiErr.Code() {
			case http.StatusNotFound:
				return filters.ErrDimensionOptionsNotFound
			}
		}
		return err
	}

	return nil
}

func (api *FilterAPI) createFilterOutputResource(ctx context.Context, newFilter *models.Filter, filterBlueprintID string) (models.Filter, error) {
	filterOutput := *newFilter
	u, err := uuid.NewV4()
	if err != nil {
		log.Event(ctx, "failed to create a new UUID for i", log.ERROR, log.Error(err))
		return models.Filter{}, err
	}
	filterOutput.FilterID = u.String()
	filterOutput.State = models.CreatedState
	filterOutput.Links.Self.HRef = fmt.Sprintf("%s/filter-outputs/%s", api.host, filterOutput.FilterID)
	filterOutput.Links.Dimensions.HRef = ""
	filterOutput.Links.FilterBlueprint.HRef = fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID)
	filterOutput.Links.FilterBlueprint.ID = filterBlueprintID
	filterOutput.LastUpdated = time.Now()

	// Clear out any event information to output document
	filterOutput.Events = []*models.Event{
		{
			Type: eventFilterOutputCreated,
			Time: time.Now(),
		},
	}

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
		log.Event(ctx, "unable to create filter output", log.ERROR, log.Error(err), log.Data{"filter_output": filterOutput})
		return models.Filter{}, err
	}

	log.Event(ctx, "submitting filter job", log.INFO, log.Data{"filter_id": filterOutput.FilterID})
	return filterOutput, api.outputQueue.Queue(&filterOutput)
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setJSONPatchContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json-patch+json")
}

func setETag(w http.ResponseWriter, eTag string) {
	w.Header().Set("ETag", eTag)
}

func getIfMatch(r *http.Request) string {
	return r.Header.Get("If-Match")
}

func getIfMatchForce(r *http.Request) (string, error) {
	eTag := getIfMatch(r)
	if eTag == "" {
		err := filters.ErrNoIfMatchHeader
		return "", err
	}
	return eTag, nil
}

func createNewFilter(filter *models.Filter, currentFilter *models.Filter) (newFilter *models.Filter, versionHasChanged bool) {
	newFilter = currentFilter

	if filter.Dataset != nil {
		if filter.Dataset.Version != 0 && filter.Dataset.Version != currentFilter.Dataset.Version {
			versionHasChanged = true
			newFilter.Dataset.Version = filter.Dataset.Version
		}

		if &filter.Events != nil {
			newFilter.Events = append(newFilter.Events, filter.Events...)
		}
	}

	return
}

func setErrorCode(w http.ResponseWriter, err error, typ ...string) {

	switch err {
	case filters.ErrFilterBlueprintNotFound:
		if typ != nil && typ[0] == statusBadRequest {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case filters.ErrDimensionNotFound:
		if typ != nil && typ[0] == statusBadRequest {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case filters.ErrDimensionsNotFound:
		fallthrough
	case filters.ErrVersionNotFound:
		if typ != nil {
			if typ[0] == statusBadRequest {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if typ[0] == statusUnprocessableEntity {
				http.Error(w, "version for filter blueprint no longer exists", http.StatusUnprocessableEntity)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case filters.ErrDimensionOptionNotFound:
		fallthrough
	case filters.ErrFilterOutputNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case filters.ErrUnauthorised:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case filters.ErrInvalidQueryParameter:
		http.Error(w, filters.ErrInvalidQueryParameter.Error(), http.StatusBadRequest)
		return
	case filters.ErrBadRequest:
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	case filters.ErrFilterBlueprintConflict:
		http.Error(w, err.Error(), http.StatusConflict)
	case filters.ErrFilterOutputConflict:
		http.Error(w, err.Error(), http.StatusConflict)
	case filters.ErrInternalError:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	default:

		switch err.(type) {
		case filters.BadRequestErr:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case filters.ForbiddenErr:
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		default:
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}
	}
}
