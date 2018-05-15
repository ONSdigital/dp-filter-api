package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"

	"fmt"

	"strconv"

	"github.com/ONSdigital/go-ns/common"
	"github.com/satori/go.uuid"
)

var (
	internalError = "Failed to process the request due to an internal error"
	badRequest    = "Bad request - Invalid request body"
	unauthorised  = "Unauthorised, request lacks valid authentication credentials"
	forbidden     = "Forbidden, the filter output has been locked as it has been submitted to be processed"

	statusBadRequest          = "Bad request"
	statusUnprocessableEntity = "Unprocessable entity"

	incorrectDimensionOptions = regexp.MustCompile("Bad request - incorrect dimension options chosen")
	incorrectDimension        = regexp.MustCompile("Bad request - incorrect dimensions chosen")

	errNotFound              = errors.New("Not found")
	errForbidden             = errors.New("Forbidden")
	errAuth                  = errors.New(unauthorised)
	errNoAuthHeader          = errors.New("No auth header provided")
	errDimensionBadRequest   = errors.New("Bad request - filter dimension not found")
	errDimensionNotFound     = errors.New("Dimension not found")
	errOptionNotFound        = errors.New("Option not found")
	errFilterOutputNotFound  = errors.New("Filter output not found")
	errRequestLimitNotNumber = errors.New("requested limit is not a number")
	errMissingDimensions     = errors.New("missing dimensions")
)

const filterSubmitted = "true"

func (api *FilterAPI) addFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	submitted := r.FormValue("submitted")
	logData := log.Data{"submitted": submitted}
	log.Info("create filter blueprint", logData)

	newFilter := models.Filter{}
	filterParameters, err := models.CreateNewFilter(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err = filterParameters.ValidateNewFilter(); err != nil {
		logData["filter_parameters"] = filterParameters
		log.ErrorC("filter parameters falied validation", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// Create unique id
	newFilter.FilterID = uuid.NewV4().String()
	newFilter.Dimensions = filterParameters.Dimensions
	logData["new_filter"] = newFilter

	// add version information from datasetAPI
	version, err := api.datasetAPI.GetVersion(r.Context(), *filterParameters.Dataset)
	if err != nil {
		log.ErrorC("unable to retrieve version document", err, logData)
		setErrorCode(w, err)
		return
	}

	if version.State != publishedState && !common.IsCallerPresent(r.Context()) {
		log.Info("unauthenticated request to filter unpublished version", log.Data{"dataset": *filterParameters.Dataset, "state": version.State})
		http.Error(w, badRequest, http.StatusBadRequest)
		return
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
	newFilter.Dataset = filterParameters.Dataset
	logData["new_filter"] = newFilter

	if err = api.checkFilterOptions(r.Context(), &newFilter, version); err != nil {
		log.ErrorC("failed to select valid filter options", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = api.dataStore.AddFilter(api.host, &newFilter); err != nil {
		log.ErrorC("failed to create new filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	if submitted == filterSubmitted {
		var filterOutput models.Filter
		// Create filter output resource and use id to pass into kafka
		filterOutput, err = api.createFilterOutputResource(&newFilter, newFilter.FilterID)
		if err != nil {
			log.ErrorC("failed to create new filter output", err, logData)
			setErrorCode(w, err)
			return
		}
		logData["filter_output_id"] = filterOutput.FilterID

		newFilter.Links = links
		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID

		logData["new_filter"] = newFilter

		log.Info("filter output id sent in message to kafka", logData)
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

	log.Info("created filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	log.Info("getting filter blueprint", logData)

	filterBlueprint, err := api.getFilter(r.Context(), filterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	filterBlueprint.Dimensions = nil
	logData["filter_blueprint"] = filterBlueprint

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint into bytes", err, logData)
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

	log.Info("got filter blueprint", logData)
}

func (api *FilterAPI) updateFilterBlueprint(w http.ResponseWriter, r *http.Request) {
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

	currentFilter, err := api.getFilter(r.Context(), filterID)
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

func (api *FilterAPI) getFilter(ctx context.Context, filterID string) (*models.Filter, error) {
	filter, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		return nil, err
	}

	//only return the filter if it is for published data or via authenticated request
	if filter.Published != nil && *filter.Published == models.Published || common.IsCallerPresent(ctx) {
		return filter, nil
	}

	log.Info("unauthenticated request to access unpublished filter", log.Data{"filter_blueprint": filter})

	version, err := api.datasetAPI.GetVersion(ctx, *filter.Dataset)
	if err != nil {
		log.Error(errors.New("failed to retrieve version from dataset api"), log.Data{"filter_blueprint": filter})
		return nil, errNotFound
	}

	//version has been published since filter was last requested, so update filter and return
	if version.State == publishedState {
		filter.Published = &models.Published
		if err := api.dataStore.UpdateFilter(filter); err != nil {
			log.Error(err, log.Data{"filter_id": filterID})
			return nil, errNotFound
		}

		return filter, nil
	}

	return nil, errNotFound
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
		err = fmt.Errorf("Bad request - incorrect dimension options chosen: %v", incorrectDimensionOptions)
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
	switch err.Error() {
	case "Not found":
		if typ != nil && typ[0] == statusBadRequest {
			http.Error(w, "Bad request - filter blueprint not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Filter blueprint not found", http.StatusNotFound)
		return
	case "Dimension not found":
		if typ != nil && typ[0] == statusBadRequest {
			http.Error(w, "Bad request - dimension not found", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case "Option not found":
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case "Filter output not found":
		http.Error(w, err.Error(), http.StatusNotFound)
	case "Version not found":
		if typ != nil {
			if typ[0] == statusBadRequest {
				http.Error(w, "Bad request - version not found", http.StatusBadRequest)
				return
			}
			if typ[0] == statusUnprocessableEntity {
				http.Error(w, "Unprocessable entity - version for filter blueprint no longer exists", http.StatusUnprocessableEntity)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case "Bad request - filter blueprint not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case "Bad request - filter dimension not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case "Bad request - filter or dimension not found":
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case "Bad request":
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	case errNoAuthHeader.Error():
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	case errAuth.Error():
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
