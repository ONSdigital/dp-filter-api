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

	uuid "github.com/satori/go.uuid"
)

var (
	internalError             = "Failed to process the request due to an internal error"
	badRequest                = "Bad request - Invalid request body"
	unauthorised              = "Unauthorised, request lacks valid authentication credentials"
	statusBadRequest          = "Bad request"
	statusUnprocessableEntity = "Unprocessable entity"
	incorrectDimensionOptions = regexp.MustCompile("Bad request - incorrect dimension options chosen")
	incorrectDimension        = regexp.MustCompile("Dimension not found")
	errAuth                   = errors.New(unauthorised)
	errNoAuthHeader           = errors.New("No auth header provided")
	errRequestLimitNotNumber  = errors.New("requested limit is not a number")
	errMissingDimensions      = errors.New("missing dimensions")
)

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

	_, err = api.dataStore.AddFilter(api.host, &newFilter)
	if err != nil {
		log.ErrorC("failed to create new filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	if submitted == "true" {

		// Create filter output resource and use id to pass into kafka
		filterOutput, err := api.createFilterOutputResource(&newFilter, newFilter.FilterID)
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

	log.Info("created new filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	log.Info("getting filter blueprint", logData)

	filterBlueprint, err := api.dataStore.GetFilter(filterID)
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

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	logData := log.Data{"filter_blueprint_id": filterID}
	log.Info("getting filter blueprint dimensions", logData)

	dimensions, err := api.dataStore.GetFilterDimensions(filterID)
	if err != nil {
		log.ErrorC("unable to get dimensions for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["dimensions"] = dimensions

	bytes, err := json.Marshal(dimensions)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint dimensions into bytes", err, logData)
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

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("getting filter blueprint dimension", logData)

	err := api.dataStore.GetFilterDimension(filterID, name)
	if err != nil {
		log.ErrorC("unable to get dimension from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) removeFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("removing filter blueprint dimension", logData)

	err := api.dataStore.RemoveFilterDimension(filterID, name)
	if err != nil {
		log.ErrorC("unable to remove dimension from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("add filter blueprint dimension", logData)

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	addDimension := &models.AddDimension{
		FilterID: filterID,
		Name:     name,
		Options:  options,
	}
	logData["dimension_options"] = addDimension

	// get filter blueprint to retreive the version for an edition of a dataset
	filterBlueprint, err := api.dataStore.GetFilter(addDimension.FilterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["current_filter_blueprint"] = filterBlueprint

	if err = api.checkNewFilterDimension(r.Context(), addDimension, *filterBlueprint.Dataset); err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		if err == ErrVersionNotFound {
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = api.dataStore.AddFilterDimension(addDimension); err != nil {
		log.ErrorC("failed to add dimension to filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension for filter blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
	}
	log.Info("get filter blueprint dimension options", logData)

	dimensionOptions, err := api.dataStore.GetFilterDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorC("unable to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["dimension_options"] = dimensionOptions

	bytes, err := json.Marshal(dimensionOptions)
	if err != nil {
		log.ErrorC("failed to marshal filter blueprint dimension options into bytes", err, logData)
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

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
		"option":              option,
	}
	log.Info("get filter blueprint dimension option", logData)

	err := api.dataStore.GetFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.ErrorC("unable to get dimension option for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
		"option":              option,
	}
	log.Info("add filter blueprint dimension option", logData)

	addDimensionOption := &models.AddDimensionOption{
		FilterID: filterID,
		Name:     name,
		Option:   option,
	}

	// get filter blueprint to retreive version id
	filterBlueprint, err := api.dataStore.GetFilter(addDimensionOption.FilterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		setErrorCode(w, err, statusBadRequest)
		return
	}
	logData["current_filter_blueprint"] = filterBlueprint

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	addDimensionOptions := &models.AddDimension{
		Name:    addDimensionOption.Name,
		Options: []string{addDimensionOption.Option},
	}
	logData["dimension_options"] = addDimensionOptions

	if err = api.checkNewFilterDimension(r.Context(), addDimensionOptions, *filterBlueprint.Dataset); err != nil {

		if err == ErrVersionNotFound {
			log.ErrorC("failed to select valid version", err, logData)
			setErrorCode(w, err, statusUnprocessableEntity)
			return
		}

		if incorrectDimensionOptions.MatchString(err.Error()) {
			log.ErrorC("failed to select valid filter dimension options", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if incorrectDimension.MatchString(err.Error()) {
			log.ErrorC("failed to select valid filter dimension", err, logData)
			setErrorCode(w, err, statusBadRequest)
			return
		}

		log.ErrorC("failed to successfully check filter dimensions", err, logData)
		setErrorCode(w, err)
		return
	}

	if err := api.dataStore.AddFilterDimensionOption(addDimensionOption); err != nil {
		log.ErrorC("failed to add dimension option to filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", logData)
}

func (api *FilterAPI) updateFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	submitted := r.URL.Query().Get("submitted")
	logData := log.Data{"filter_blueprint_id": filterID, "submitted": submitted}
	log.Info("updating filter blueprint", logData)

	const FilterSubmitted = "true"
	filter, err := models.CreateFilter(r.Body)
	if err != nil {

		// When filter blueprint has query parameter `submitted` set to true then
		// request can have an empty json in body for this PUT request
		if submitted != FilterSubmitted || err != models.ErrorNoData {
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

	currentFilter, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.ErrorC("unable to get filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}
	logData["current_filter"] = currentFilter

	newFilter, versionHasChanged := createNewFilter(filter, currentFilter)
	logData["new_filter"] = newFilter

	if versionHasChanged {
		log.Info("finding instance id for filter after version change", logData)
		// add version information from datasetAPI for new version
		version, err := api.datasetAPI.GetVersion(r.Context(), *newFilter.Dataset)
		if err != nil {
			log.ErrorC("unable to retrieve version document", err, logData)
			setErrorCode(w, err, statusBadRequest)
			return
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

	if submitted == FilterSubmitted {
		outputFilter := newFilter

		// Create filter output resource and use id to pass into kafka
		filterOutput, err := api.createFilterOutputResource(outputFilter, filterID)
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

func (api *FilterAPI) getFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.Info("getting filter output", logData)

	filterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
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

	log.Info("got filtered output", logData)
}

func (api *FilterAPI) updateFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	logData := log.Data{"filter_output_id": filterOutputID}
	log.Info("updating filter output", logData)

	if r.Context().Value(internalToken) != true {
		err := errors.New("Not authorised")
		log.ErrorC("failed to update filter output", err, logData)
		setErrorCode(w, errNoAuthHeader)
		return
	}

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.ErrorC("unable to unmarshal request body", err, logData)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}
	logData["filter_output"] = filterOutput

	if err = filterOutput.ValidateFilterOutputUpdate(); err != nil {
		log.ErrorC("filter output failed validation", err, logData)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	filterOutput.FilterID = filterOutputID

	// TODO check filter output resource for current downloads
	previousFilterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.ErrorC("unable to get current filter output", err, logData)
		setErrorCode(w, err)
		return
	}

	updatedFilterOutput := checkFilterOutputQuery(previousFilterOutput, filterOutput)
	logData["filter_output_update"] = updatedFilterOutput

	if err = api.dataStore.UpdateFilterOutput(updatedFilterOutput); err != nil {
		log.ErrorC("unable to update filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("got filtered blueprint", logData)
}

func (api *FilterAPI) createFilterOutputResource(newFilter *models.Filter, filterBlueprintID string) (models.Filter, error) {
	filterOutput := *newFilter
	filterOutput.FilterID = uuid.NewV4().String()
	filterOutput.State = models.CreatedState
	filterOutput.Links.Self.HRef = fmt.Sprintf("%s/filter-outputs/%s", api.host, filterOutput.FilterID)
	filterOutput.Links.Dimensions.HRef = ""
	filterOutput.Links.FilterBlueprint.HRef = fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID)
	filterOutput.Links.FilterBlueprint.ID = filterBlueprintID

	// Clear out any event information to output document
	filterOutput.Events = models.Events{}

	// Downloads object should exist for filter output resource
	// even if it they are empty
	filterOutput.Downloads = &models.Downloads{
		CSV:  models.DownloadItem{},
		XLS:  models.DownloadItem{},
		JSON: models.DownloadItem{},
	}

	// Remove dimension url from output filter resource
	for i := range newFilter.Dimensions {
		filterOutput.Dimensions[i].URL = ""
	}

	err := api.dataStore.CreateFilterOutput(&filterOutput)
	if err != nil {
		log.ErrorC("unable to create filter output", err, log.Data{"filter_output": filterOutput})
		return models.Filter{}, err
	}

	return filterOutput, api.outputQueue.Queue(&filterOutput)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterID,
		"dimension":           name,
		"option":              option,
	}
	log.Info("remove filter blueprint dimension option", logData)

	err := api.dataStore.RemoveFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.ErrorC("unable to remove dimension option from filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", logData)
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

	filterOutput, err := api.dataStore.GetFilterOutput(filterID)
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
		// Call dimension options list endpoint
		datasetDimensionOptions, err := api.datasetAPI.GetVersionDimensionOptions(ctx, *newFilter.Dataset, filterDimension.Name)
		if err != nil {
			logData["dimension"] = filterDimension
			log.ErrorC("failed to retreive a list of dimension options from dataset API", err, logData)
			return err
		}
		logData["dimension_options"] = datasetDimensionOptions

		log.Info("dimension options retreived from dataset API", logData)

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

func (api *FilterAPI) checkNewFilterDimension(ctx context.Context, newDimension *models.AddDimension, dataset models.Dataset) error {
	logData := log.Data{"new_dimension": newDimension, "dataset": dataset}
	log.Info("check filter dimensions and dimension options before calling api, see version number", logData)

	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.datasetAPI.GetVersionDimensions(ctx, dataset)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimensions from the dataset API", err, logData)
		return err
	}

	dimension := models.Dimension{
		Name:    newDimension.Name,
		Options: newDimension.Options,
	}

	if err = models.ValidateFilterDimensions([]models.Dimension{dimension}, datasetDimensions); err != nil {
		log.ErrorC("filter dimensions failed validation", err, logData)
		return err
	}

	// Call dimension options endpoint
	datasetDimensionOptions, err := api.datasetAPI.GetVersionDimensionOptions(ctx, dataset, dimension.Name)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimension options from the dataset API", err, logData)
		return err
	}

	var incorrectDimensionOptions []string
	incorrectOptions := models.ValidateFilterDimensionOptions(dimension.Options, datasetDimensionOptions)
	if incorrectOptions != nil {
		incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
	}

	if incorrectDimensionOptions != nil {
		err = fmt.Errorf("Bad request - incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.ErrorC("incorrect dimension options chosen", err, logData)
		return err
	}

	return nil
}

func checkFilterOutputQuery(previousFilterOutput, filterOutput *models.Filter, typ ...string) *models.Filter {
	if previousFilterOutput.Downloads == nil {
		return filterOutput
	}

	if previousFilterOutput.Downloads.CSV.URL != "" {
		filterOutput.Downloads.CSV = previousFilterOutput.Downloads.CSV
	}

	if previousFilterOutput.Downloads.XLS.URL != "" {
		filterOutput.Downloads.XLS = previousFilterOutput.Downloads.XLS
	}

	if previousFilterOutput.Downloads.JSON.URL != "" {
		filterOutput.Downloads.JSON = previousFilterOutput.Downloads.JSON
	}

	return filterOutput
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
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
