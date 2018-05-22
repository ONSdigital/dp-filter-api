package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"fmt"

	"strconv"

	"github.com/satori/go.uuid"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/dp-filter-api/filters"
	"regexp"
)

var (
	internalError = "Failed to process the request due to an internal error"
	badRequest    = "Bad request - Invalid request body"
	forbidden     = "Forbidden, the filter output has been locked as it has been submitted to be processed"

	statusBadRequest          = "bad request"
	statusUnprocessableEntity = "unprocessable entity"

	incorrectDimensionOptions = regexp.MustCompile("incorrect dimension options chosen")
	incorrectDimension        = regexp.MustCompile("incorrect dimensions chosen")

	errForbidden    = errors.New("forbidden")
)

const (
	filterSubmitted = "true"

	// audit actions
	createFilterBlueprintAction = "createFilterBlueprint"
	getFilterBlueprintAction = "getFilterBlueprint"

	// audit results
	actionAttempted    = "attempted"
	actionSuccessful   = "successful"
	actionUnsuccessful = "unsuccessful"
)

func (api *FilterAPI) postFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	submitted := r.FormValue("submitted")
	logData := log.Data{"submitted": submitted}
	log.Info("create filter blueprint", logData)

	if auditErr := api.auditor.Record(r.Context(), createFilterBlueprintAction, actionAttempted, nil); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	filter, err := models.CreateNewFilter(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		if auditErr := api.auditor.Record(r.Context(), createFilterBlueprintAction, actionUnsuccessful, nil); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	newFilter, err := api.createFilterBlueprint(filter, submitted, r.Context())
	if err != nil {
		log.ErrorC("failed to create new filter", err, logData)
		if auditErr := api.auditor.Record(r.Context(), createFilterBlueprintAction, actionUnsuccessful, nil); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	log.Info("created filter blueprint", logData)
	if auditErr := api.auditor.Record(r.Context(), createFilterBlueprintAction, actionSuccessful, nil); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint into bytes", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.ErrorC("failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}
}

func (api *FilterAPI) createFilterBlueprint(filter *models.NewFilter, submitted string, ctx context.Context) (*models.Filter, error) {

	newFilter := &models.Filter{}
	logData := log.Data{}

	if err := filter.ValidateNewFilter(); err != nil {
		logData["filter_parameters"] = filter
		log.ErrorC("filter parameters failed validation", err, logData)
		return nil, filters.ErrBadRequest
	}

	// Create unique id
	newFilter.FilterID = uuid.NewV4().String()
	newFilter.Dimensions = filter.Dimensions
	logData["new_filter"] = newFilter

	// add version information from datasetAPI
	version, err := api.datasetAPI.GetVersion(ctx, *filter.Dataset)
	if err != nil {
		log.ErrorC("unable to retrieve version document", err, logData)
		return nil, err
	}

	if version.State != publishedState && !common.IsCallerPresent(ctx) {
		log.Info("unauthenticated request to filter unpublished version", log.Data{"dataset": *filter.Dataset, "state": version.State})
		return nil, filters.ErrBadRequest
	}

	if version.State == publishedState {
		newFilter.Published = &models.Published
	}

	links := models.LinkMap{
		Dimensions: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s/dimensions", api.host, newFilter.FilterID),
		},
		Self: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s", api.host, newFilter.FilterID),
		},
		Version: models.LinkObject{
			HRef: version.Links.Self.HRef,
			ID:   strconv.Itoa(version.Version),
		},
	}

	newFilter.Links = links
	newFilter.InstanceID = version.ID
	newFilter.Dataset = filter.Dataset
	logData["new_filter"] = newFilter

	if err = api.checkFilterOptions(ctx, newFilter, version); err != nil {
		log.ErrorC("failed to select valid filter options", err, logData)
		return nil, filters.NewBadRequestErr(err.Error())
	}

	if _, err = api.dataStore.AddFilter(api.host, newFilter); err != nil {
		log.ErrorC("failed to create new filter blueprint", err, logData)
		return nil, err
	}

	if submitted == filterSubmitted {
		var filterOutput models.Filter
		// Create filter output resource and use id to pass into kafka
		filterOutput, err = api.createFilterOutputResource(newFilter, newFilter.FilterID)
		if err != nil {
			log.ErrorC("failed to create new filter output", err, logData)
			return nil, err
		}
		logData["filter_output_id"] = filterOutput.FilterID

		newFilter.Links = links
		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID

		logData["new_filter"] = newFilter
		log.Info("filter output id sent in message to kafka", logData)
	}

	return newFilter, nil
}

func (api *FilterAPI) getFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	log.Info("getting filter blueprint", logData)

	auditParams := common.Params{"filter_blueprint_id": filterID}
	if auditErr := api.auditor.Record(r.Context(), getFilterBlueprintAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	filterBlueprint, err := api.getFilterBlueprint(r.Context(), filterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterBlueprintAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}

		setErrorCode(w, err)
		return
	}

	filterBlueprint.Dimensions = nil
	logData["filter_blueprint"] = filterBlueprint

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint into bytes", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getFilterBlueprintAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	log.Info("got filter blueprint", logData)
	if auditErr := api.auditor.Record(r.Context(), getFilterBlueprintAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.ErrorC("failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
	}
}

func (api *FilterAPI) putFilterBlueprintHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	submitted := r.URL.Query().Get("submitted")
	logData := log.Data{"filter_blueprint_id": filterID, "submitted": submitted}
	log.Info("updating filter blueprint", logData)

	filter, err := models.CreateFilter(r.Body)
	if err != nil {
		// When filter blueprint has query parameter `submitted` set to true then
		// request can have an empty json in body for this PUT request
		if submitted != filterSubmitted || err != models.ErrorNoData {
			log.ErrorC("unable to unmarshal request body", err, logData)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
	}

	filter.FilterID = filterID
	logData["filter_update"] = filter

	if err = models.ValidateFilterBlueprintUpdate(filter); err != nil {
		log.ErrorC("filter blueprint failed validation", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	currentFilter, err := api.getFilterBlueprint(r.Context(), filterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	logData["current_filter"] = currentFilter

	newFilter, versionHasChanged := createNewFilter(filter, currentFilter)
	logData["new_filter"] = newFilter

	if versionHasChanged {
		log.Info("finding new version details for filter after version change", logData)

		var version *models.Version
		// add version information from datasetAPI for new version
		version, err = api.datasetAPI.GetVersion(r.Context(), *newFilter.Dataset)
		if err != nil {
			log.ErrorC("unable to retrieve version document", err, logData)
			setErrorCode(w, err, statusBadRequest)
			return
		}

		newFilter.Published = &models.Unpublished
		if version.State == "published" {
			newFilter.Published = &models.Published
		}

		newFilter.InstanceID = version.ID
		newFilter.Links.Version.HRef = version.Links.Self.HRef

		// Check existing dimensions work for new version
		if err = api.checkFilterOptions(r.Context(), newFilter, version); err != nil {
			log.ErrorC("failed to select valid filter options", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	err = api.dataStore.UpdateFilter(newFilter)
	if err != nil {
		log.ErrorC("unable to update filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	if submitted == filterSubmitted {
		outputFilter := newFilter

		var filterOutput models.Filter
		// Create filter output resource and use id to pass into kafka
		filterOutput, err = api.createFilterOutputResource(outputFilter, filterID)
		if err != nil {
			log.ErrorC("failed to create new filter output", err, logData)
			setErrorCode(w, err)
			return
		}
		logData["filter_output_id"] = filterOutput.FilterID

		log.Info("filter output id sent in message to kafka", logData)

		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID
	}

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.ErrorC("failed to marshal updated filter blueprint into bytes", err, logData)
		setErrorCode(w, err)
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

	log.Info("filter blueprint updated", logData)
}

func (api *FilterAPI) getFilterBlueprint(ctx context.Context, filterID string) (*models.Filter, error) {

	logData := log.Data{"filter_blueprint_id": filterID}

	filter, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, logData)
		return nil, err
	}

	//only return the filter if it is for published data or via authenticated request
	if filter.Published != nil && *filter.Published == models.Published || common.IsCallerPresent(ctx) {
		return filter, nil
	}

	log.Info("unauthenticated request to access unpublished filter", logData)

	version, err := api.datasetAPI.GetVersion(ctx, *filter.Dataset)
	if err != nil {
		log.Error(errors.New("failed to retrieve version from dataset api"), logData)
		return nil, err
	}

	//version has been published since filter was last requested, so update filter and return
	if version.State == publishedState {
		filter.Published = &models.Published
		if err := api.dataStore.UpdateFilter(filter); err != nil {
			log.Error(err, logData)
			return nil, filters.ErrFilterBlueprintNotFound
		}

		return filter, nil
	}

	// not authenticated, so return not found
	return nil, filters.ErrFilterBlueprintNotFound
}

func (api *FilterAPI) checkFilterOptions(ctx context.Context, newFilter *models.Filter, version *models.Version) error {
	logData := log.Data{"new_filter": newFilter, "version": version.Version}
	log.Info("check filter dimension options before calling api, see version number", logData)

	// Call dimensions list endpoint
	datasetDimensions, err := api.datasetAPI.GetVersionDimensions(ctx, *newFilter.Dataset)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimensions from the dataset API", err, logData)
		return err
	}
	logData["dataset_dimensions"] = datasetDimensions

	log.Info("dimensions retreived from dataset API", logData)

	if err = models.ValidateFilterDimensions(newFilter.Dimensions, datasetDimensions); err != nil {
		log.ErrorC("filter dimensions failed validation", err, logData)
		return err
	}
	log.Info("successfully validated filter dimensions", logData)

	var incorrectDimensionOptions []string
	for _, filterDimension := range newFilter.Dimensions {
		localData := logData

		var datasetDimensionOptions *models.DatasetDimensionOptionResults
		// Call dimension options list endpoint
		datasetDimensionOptions, err = api.datasetAPI.GetVersionDimensionOptions(ctx, *newFilter.Dataset, filterDimension.Name)
		if err != nil {
			localData["dimension"] = filterDimension
			log.ErrorC("failed to retreive a list of dimension options from dataset API", err, localData)
			return err
		}
		localData["dimension_options"] = datasetDimensionOptions

		log.Info("dimension options retreived from dataset API", localData)

		incorrectOptions := models.ValidateFilterDimensionOptions(filterDimension.Options, datasetDimensionOptions)
		if incorrectOptions != nil {
			incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
		}
	}

	if incorrectDimensionOptions != nil {
		logData["incorrect_dimension_options"] = incorrectDimensionOptions
		err = fmt.Errorf("incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.ErrorC("incorrect dimension options chosen", err, logData)
		return err
	}

	return nil
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func createNewFilter(filter *models.Filter, currentFilter *models.Filter) (newFilter *models.Filter, versionHasChanged bool) {
	newFilter = currentFilter

	if filter.Dataset != nil {
		if filter.Dataset.Version != 0 && filter.Dataset.Version != currentFilter.Dataset.Version {
			versionHasChanged = true
			newFilter.Dataset.Version = filter.Dataset.Version
		}

		if &filter.Events != nil {
			if filter.Events.Info != nil {
				newFilter.Events.Info = append(newFilter.Events.Info, filter.Events.Info...)
			}

			if filter.Events.Error != nil {
				newFilter.Events.Error = append(newFilter.Events.Error, filter.Events.Error...)
			}
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
	case filters.ErrOptionNotFound:
		fallthrough
	case filters.ErrFilterOutputNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case filters.ErrUnauthorised:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case filters.ErrBadRequest:
		http.Error(w, badRequest, http.StatusBadRequest)
		return

	default:

		switch err.(type) {
		case filters.BadRequestErr:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}
	}
}

func handleAuditingFailure(w http.ResponseWriter, err error, logData log.Data) {
	log.ErrorC("error while attempting to record audit event, failing request", err, logData)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
