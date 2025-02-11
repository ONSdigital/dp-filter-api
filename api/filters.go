package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	datasetAPI "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-net/v2/links"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

var (
	InternalError = "Failed to process the request due to an internal error"
	BadRequest    = "Bad request - Invalid request body"

	statusBadRequest          = "bad request"
	statusUnprocessableEntity = "unprocessable entity"

	incorrectDimensionOptions = regexp.MustCompile("incorrect dimension options chosen")
	incorrectDimension        = regexp.MustCompile("incorrect dimensions chosen")

	publishedState = "published"
)

const (
	filterSubmitted = "true"
)

func (api *FilterAPI) postFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	submitted := r.FormValue("submitted")
	logData := log.Data{"submitted": submitted}
	ctx := r.Context()
	log.Info(ctx, "create filter blueprint", logData)

	filter, err := models.CreateNewFilter(r.Body)
	if err != nil {
		log.Error(ctx, "unable to unmarshal request body", err, logData)
		if err, ok := err.(models.DuplicateDimensionError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, BadRequest, http.StatusBadRequest)
		}
		return
	}

	newFilter, err := api.createFilterBlueprint(ctx, filter, submitted)
	if err != nil {
		log.Error(ctx, "failed to create new filter", err, logData)
		setErrorCode(w, err)
		return
	}
	log.Info(ctx, "created filter blueprint", logData)

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint into bytes", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newFilter.ETag)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}
}

func (api *FilterAPI) createFilterBlueprint(ctx context.Context, filter *models.NewFilter, submitted string) (*models.Filter, error) {

	newFilter := &models.Filter{}
	logData := log.Data{}

	if err := filter.ValidateNewFilter(); err != nil {
		logData["filter_parameters"] = filter
		log.Error(ctx, "filter parameters failed validation", err, logData)
		return nil, filters.ErrBadRequest
	}

	// Create unique id
	u, err := uuid.NewV4()
	if err != nil {
		log.Error(ctx, "failed to create a new UUID for filter blueprint", err, logData)
		return nil, err
	}
	newFilter.FilterID = u.String()
	newFilter.Dimensions = filter.Dimensions
	logData["new_filter"] = newFilter

	// add version information from datasetAPI
	version, err := api.getVersion(ctx, filter.Dataset)
	if err != nil {
		log.Error(ctx, "unable to retrieve version document", err, logData)
		return nil, err
	}

	if version.State != publishedState && !dprequest.IsCallerPresent(ctx) {
		log.Info(ctx, "unauthenticated request to filter unpublished version", log.Data{"dataset": *filter.Dataset, "state": version.State})
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
		log.Error(ctx, "failed to select valid filter options", err, logData)
		return nil, filters.NewBadRequestErr(err.Error())
	}

	newFilter, err = api.dataStore.AddFilter(ctx, newFilter)
	if err != nil {
		log.Error(ctx, "failed to create new filter blueprint", err, logData)
		return nil, err
	}

	if submitted == filterSubmitted {
		var filterOutput models.Filter
		// Create filter output resource and use filter id, dataset, edition and version to pass into kafka
		filterOutput, err = api.createFilterOutputResource(ctx, newFilter, newFilter.FilterID)
		if err != nil {
			log.Error(ctx, "failed to create new filter output", err, logData)
			return nil, err
		}
		logData["filter_output_id"] = filterOutput.FilterID

		newFilter.Links = links
		newFilter.Links.FilterOutput = &models.LinkObject{
			HRef: filterOutput.Links.Self.HRef,
			ID:   filterOutput.FilterID,
		}

		logData["new_filter"] = newFilter
		log.Info(ctx, "filter output id sent in message to kafka", logData)
	}

	return newFilter, nil
}

func (api *FilterAPI) getFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	ctx := r.Context()
	log.Info(ctx, "getting filter blueprint", logData)

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "unable to get filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	filterBlueprint.ID = filterBlueprint.FilterID
	filterBlueprint.Dimensions = nil
	logData["filter_blueprint"] = filterBlueprint

	enableURLRewriting := true

	if enableURLRewriting {
		dimensionSearchAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)

		fmt.Println("host is : ", api.host)
		linkFields := []*models.LinkObject{
			filterBlueprint.Links.Dimensions,
			filterBlueprint.Links.FilterOutput,
			filterBlueprint.Links.FilterBlueprint,
			filterBlueprint.Links.Self,
			filterBlueprint.Links.Version,
		}

		for _, linkObj := range linkFields {
			if linkObj != nil {
				newLink, err := dimensionSearchAPILinksBuilder.BuildLink(linkObj.HRef)
				if err == nil {
					linkObj.HRef = newLink
				}
			}
		}
	}

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filterBlueprint.ETag)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
	}

	log.Info(ctx, "got filter blueprint", logData)
}

func (api *FilterAPI) postFilterBlueprintSubmitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
	}
	ctx := r.Context()
	log.Info(ctx, "post filter blueprint", logData)

	err := errors.New("filter not of type flexible")
	log.Error(ctx, "invalid filter type", err, logData)
	http.Error(w, BadRequest, http.StatusBadRequest)
}

func (api *FilterAPI) putFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	defer dphttp.DrainBody(r)

	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	submitted := r.URL.Query().Get("submitted")
	logData := log.Data{"filter_blueprint_id": filterID, "submitted": submitted}
	ctx := r.Context()
	log.Info(ctx, "updating filter blueprint", logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		// When filter blueprint has query parameter `submitted` set to true then
		// request can have an empty json in body for this PUT request
		if submitted != filterSubmitted || err != models.ErrorNoData {
			log.Error(ctx, "unable to unmarshal request body", err, logData)
			http.Error(w, BadRequest, http.StatusBadRequest)
			return
		}
	}
	filter.FilterID = filterID

	newFilter, err := api.updateFilterBlueprint(ctx, filter, submitted, eTag)
	if err != nil {
		log.Error(ctx, "failed to update filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}
	log.Info(ctx, "filter blueprint updated", logData)

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Error(ctx, "failed to marshal updated filter blueprint into bytes", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newFilter.ETag)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}
}

func (api *FilterAPI) updateFilterBlueprint(ctx context.Context, filter *models.Filter, submitted, eTag string) (*models.Filter, error) {

	logData := log.Data{"filter_blueprint_id": filter.FilterID, "submitted": submitted}
	log.Info(ctx, "updating filter blueprint", logData)
	logData["filter_update"] = filter

	if err := models.ValidateFilterBlueprintUpdate(filter); err != nil {
		log.Error(ctx, "filter blueprint failed validation", err, logData)
		return nil, filters.ErrBadRequest
	}

	currentFilter, err := api.getFilterBlueprint(ctx, filter.FilterID, eTag)
	if err != nil {
		log.Error(ctx, "unable to get filter blueprint", err, logData)
		return nil, err
	}

	timestamp := currentFilter.UniqueTimestamp
	logData["current_filter_timestamp"] = timestamp

	logData["current_filter"] = currentFilter

	newFilter, versionHasChanged := createNewFilter(filter, currentFilter)
	logData["new_filter"] = newFilter

	if versionHasChanged {
		log.Info(ctx, "finding new version details for filter after version change", logData)

		version, err := api.getVersion(ctx, newFilter.Dataset)
		if err != nil {
			log.Error(ctx, "unable to retrieve version document", err, logData)
			return nil, filters.NewBadRequestErr(err.Error())
		}

		newFilter.Published = &models.Unpublished
		if version.State == "published" {
			newFilter.Published = &models.Published
		}

		newFilter.InstanceID = version.ID
		newFilter.Links.Version = &models.LinkObject{
			HRef: version.Links.Self.URL,
		}

		// Check existing dimensions work for new version
		if err = api.checkFilterOptions(ctx, newFilter, version); err != nil {
			log.Error(ctx, "failed to select valid filter options", err, logData)
			return nil, filters.NewBadRequestErr(err.Error())
		}
	}

	newFilter.ETag, err = api.dataStore.UpdateFilter(ctx, newFilter, timestamp, eTag, currentFilter)
	if err != nil {
		log.Error(ctx, "unable to update filter blueprint", err, logData)
		return nil, err
	}

	if submitted == filterSubmitted {
		outputFilter := newFilter

		var filterOutput models.Filter
		// Create filter output resource and use id to pass into kafka
		filterOutput, err = api.createFilterOutputResource(ctx, outputFilter, filter.FilterID)
		if err != nil {
			log.Error(ctx, "failed to create new filter output", err, logData)
			return nil, err
		}
		logData["filter_output_id"] = filterOutput.FilterID

		log.Info(ctx, "filter output id sent in message to kafka", logData)

		newFilter.Links.FilterOutput = &models.LinkObject{
			HRef: filterOutput.Links.Self.HRef,
			ID:   filterOutput.FilterID,
		}
	}

	return newFilter, nil
}

func (api *FilterAPI) getFilterBlueprint(ctx context.Context, filterID, eTag string) (*models.Filter, error) {

	logData := log.Data{"filter_blueprint_id": filterID}

	currentFilter, err := api.dataStore.GetFilter(ctx, filterID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "error getting filter", err, logData)
		return nil, err
	}

	if eTag != mongo.AnyETag && currentFilter.ETag != eTag {
		return nil, filters.ErrFilterBlueprintConflict
	}

	// only return the filter if it is for published data or via authenticated request
	if currentFilter.Published != nil && *currentFilter.Published == models.Published || dprequest.IsCallerPresent(ctx) {
		return currentFilter, nil
	}

	log.Info(ctx, "unauthenticated request to access unpublished filter", logData)

	version, err := api.getVersion(ctx, currentFilter.Dataset)
	if err != nil {
		log.Error(ctx, "failed to retrieve version from dataset api", err, logData)
		return nil, err
	}

	// version has been published since filter was last requested, so update filter and return
	if version.State == publishedState {
		filter := currentFilter
		filter.Published = &models.Published
		filter.ETag, err = api.dataStore.UpdateFilter(ctx, filter, filter.UniqueTimestamp, currentFilter.ETag, currentFilter)
		if err != nil {
			log.Error(ctx, "error updating filter", err, logData)
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
	log.Info(ctx, "check filter dimension options before calling dataset api", logData)

	// Call dimensions list endpoint
	datasetDimensions, err := api.getDimensions(ctx, newFilter.Dataset)
	if err != nil {
		log.Error(ctx, "failed to retrieve a list of dimensions from the dataset API", err, logData)
		return err
	}

	logData["dataset_dimensions_total"] = len(datasetDimensions.Items)
	if len(datasetDimensions.Items) > 30 {
		logData["dataset_dimensions_first"] = datasetDimensions.Items[0]
	} else {
		logData["dataset_dimensions"] = datasetDimensions
	}

	log.Info(ctx, "dimensions retrieved from dataset API", logData)

	if err = models.ValidateFilterDimensions(newFilter.Dimensions, datasetDimensions); err != nil {
		log.Error(ctx, "filter dimensions failed validation", err, logData)
		return err
	}
	log.Info(ctx, "successfully validated filter dimensions", logData)

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
			if apiErr.Code() == http.StatusNotFound {
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
			if apiErr.Code() == http.StatusNotFound {
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
			if apiErr.Code() == http.StatusNotFound {
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
		log.Error(ctx, "failed to create a new UUID for i", err)
		return models.Filter{}, err
	}
	filterOutput.FilterID = u.String()
	filterOutput.State = models.CreatedState
	filterOutput.Links.Self = &models.LinkObject{
		HRef: fmt.Sprintf("%s/filter-outputs/%s", api.host, filterOutput.FilterID),
	}
	filterOutput.Links.Dimensions = &models.LinkObject{
		HRef: "",
	}
	filterOutput.Links.FilterBlueprint = &models.LinkObject{
		HRef: fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID),
		ID:   filterBlueprintID,
	}
	filterOutput.LastUpdated = time.Now()

	// Clear out any event information to output document
	filterOutput.Events = []*models.Event{
		{
			Type: models.EventFilterOutputCreated,
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

	if err := api.dataStore.CreateFilterOutput(ctx, &filterOutput); err != nil {
		log.Error(ctx, "unable to create filter output", err, log.Data{"filter_output": filterOutput})
		return models.Filter{}, err
	}

	log.Info(ctx, "submitting filter job", log.Data{"filter_id": filterOutput.FilterID})
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
		http.Error(w, BadRequest, http.StatusBadRequest)
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
			http.Error(w, InternalError, http.StatusInternalServerError)
			return
		}
	}
}
