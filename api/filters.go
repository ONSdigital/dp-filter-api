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
	forbidden                 = "Forbidden, the filter output has been locked as it has been submitted to be processed"
	statusBadRequest          = "bad request"
	statusUnprocessableEntity = "unprocessable entity"

	incorrectDimensionOptions = regexp.MustCompile("Bad request - incorrect dimension options chosen")
)

const internalToken = "Internal-Token"

func (api *FilterAPI) addFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	submitted := r.FormValue("submitted")

	newFilter, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// Create unique id
	newFilter.FilterID = uuid.NewV4().String()

	if err = newFilter.ValidateFilterBlueprint(); err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	// add version information from datasetAPI
	auth := api.isAuthenticated(r.Header.Get(internalToken))
	instance, err := api.datasetAPI.GetInstance(r.Context(), newFilter.InstanceID, auth)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	if err = api.checkFilterOptions(r.Context(), newFilter, instance); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	links := models.LinkMap{
		Dimensions: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s/dimensions", api.host, newFilter.FilterID),
		},
		Self: models.LinkObject{
			HRef: fmt.Sprintf("%s/filters/%s", api.host, newFilter.FilterID),
		},
		Version: instance.Links.Version,
	}

	newFilter.Links = links

	log.Info("newFilter", log.Data{"new-filter": newFilter})

	_, err = api.dataStore.AddFilter(api.host, newFilter)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	if submitted == "true" {

		// Create filter output resource and use id to pass into kafka
		filterOutput := api.createFilterOutputResource(newFilter, newFilter.FilterID)
		log.Info("filter output id sent in message to kafka", log.Data{"filter_output_id": filterOutput.FilterID})

		newFilter.Links = links
		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID

		log.Info("newFilter 2", log.Data{"new-filter": newFilter})
	}

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"new_filter": newFilter})
		setErrorCode(w, err)
		return
	}

	log.Info("created new filter blueprint", log.Data{"filter_blueprint": newFilter})
}

func (api *FilterAPI) getFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]

	filterBlueprint, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		setErrorCode(w, err)
		return
	}

	filterBlueprint.Dimensions = nil

	bytes, err := json.Marshal(filterBlueprint)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filterBlueprint})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_id": filterID, "filter_blueprint": filterBlueprint})
}

func (api *FilterAPI) getFilterBlueprintDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]

	dimensions, err := api.dataStore.GetFilterDimensions(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensions)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimensions": dimensions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimensions": dimensions})
}

func (api *FilterAPI) getFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	err := api.dataStore.GetFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) removeFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	err := api.dataStore.RemoveFilterDimension(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) addFilterBlueprintDimension(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	options, err := models.CreateDimensionOptions(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	addDimension := &models.AddDimension{
		FilterID: vars["filter_blueprint_id"],
		Name:     vars["name"],
		Options:  options,
	}

	// get filter blueprint to retreive instance id
	filterBlueprint, err := api.dataStore.GetFilter(addDimension.FilterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": addDimension.FilterID})
		setErrorCode(w, err)
		return
	}

	// get instance to retrieve dataset id, edition and version
	auth := api.isAuthenticated(r.Header.Get(internalToken))
	instance, err := api.datasetAPI.GetInstance(r.Context(), filterBlueprint.InstanceID, auth)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": addDimension.FilterID, "instance_id": filterBlueprint.InstanceID})
		setErrorCode(w, err, statusUnprocessableEntity)
		return
	}

	if err = api.checkNewFilterDimension(r.Context(), addDimension, instance); err != nil {
		log.Error(err, nil)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = api.dataStore.AddFilterDimension(addDimension); err != nil {
		log.Error(err, log.Data{"dimension": addDimension})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension for filter blueprint", log.Data{"dimension": addDimension})
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]

	dimensionOptions, err := api.dataStore.GetFilterDimensionOptions(filterID, name)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(dimensionOptions)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension_options": dimensionOptions})
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.GetFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name, "option": option})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Info("got filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name, "option": option})
}

func (api *FilterAPI) addFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	addDimensionOption := &models.AddDimensionOption{
		FilterID: vars["filter_blueprint_id"],
		Name:     vars["name"],
		Option:   vars["option"],
	}

	// get filter blueprint to retreive instance id
	filterBlueprint, err := api.dataStore.GetFilter(addDimensionOption.FilterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": addDimensionOption.FilterID})
		setErrorCode(w, err, statusBadRequest)
		return
	}

	// get instance to retrieve dataset id, edition and version
	auth := api.isAuthenticated(r.Header.Get(internalToken))
	instance, err := api.datasetAPI.GetInstance(r.Context(), filterBlueprint.InstanceID, auth)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": addDimensionOption.FilterID, "instance_id": filterBlueprint.InstanceID})
		setErrorCode(w, err, statusUnprocessableEntity)
		return
	}

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	addDimensionOptions := &models.AddDimension{
		Name:    addDimensionOption.Name,
		Options: []string{addDimensionOption.Option},
	}

	if err = api.checkNewFilterDimension(r.Context(), addDimensionOptions, instance); err != nil {
		log.Error(err, nil)
		if incorrectDimensionOptions.MatchString(err.Error()) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		setErrorCode(w, err, statusBadRequest)
		return
	}

	if err := api.dataStore.AddFilterDimensionOption(addDimensionOption); err != nil {
		log.Error(err, log.Data{"dimension_option": addDimensionOption})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)

	log.Info("created new dimension option for filter blueprint", log.Data{"dimension_option": addDimensionOption})
}

func (api *FilterAPI) updateFilterBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	var submitted string
	submitted = r.URL.Query().Get("submitted")

	filter, err := models.CreateFilter(r.Body)
	if err != nil {

		// When filter blueprint has query parameter `submitted` set to true then
		// request can have an empty json in body for this PUT request
		if submitted != "true" || err != models.ErrorNoData {
			log.Error(err, log.Data{"filter_blueprint_id": filterID})
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
	}

	filter.FilterID = filterID

	currentFilter, err := api.dataStore.GetFilter(filterID)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filter, "filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	newFilter := currentFilter
	if filter.InstanceID != "" {

		// add version information from datasetAPI for new instance
		auth := api.isAuthenticated(r.Header.Get(internalToken))
		instance, err := api.datasetAPI.GetInstance(r.Context(), newFilter.InstanceID, auth)
		if err != nil {
			log.Error(err, log.Data{"new_filter": newFilter})
			setErrorCode(w, err, statusBadRequest)
			return
		}

		newFilter.InstanceID = filter.InstanceID
		newFilter.Links.Version.HRef = instance.Links.Version.HRef

		// Check existing dimensions work for new instance
		if err = api.checkFilterOptions(r.Context(), newFilter, instance); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if filter.Events.Error != nil {
		for _, errorEvent := range filter.Events.Error {
			newFilter.Events.Error = append(newFilter.Events.Error, errorEvent)
		}
	}

	if filter.Events.Info != nil {
		for _, infoEvent := range filter.Events.Info {
			newFilter.Events.Info = append(newFilter.Events.Info, infoEvent)
		}
	}

	err = api.dataStore.UpdateFilter(newFilter)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": filter, "filter_blueprint_id": filterID})
		setErrorCode(w, err)
		return
	}

	if submitted == "true" {
		outputFilter := newFilter

		// Create filter output resource and use id to pass into kafka
		filterOutput := api.createFilterOutputResource(outputFilter, filterID)
		log.Info("filter output id sent in message to kafka", log.Data{"filter_output_id": filterOutput.FilterID})

		newFilter.Links.FilterOutput.HRef = filterOutput.Links.Self.HRef
		newFilter.Links.FilterOutput.ID = filterOutput.FilterID
	}

	bytes, err := json.Marshal(newFilter)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": newFilter})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint": newFilter})
		setErrorCode(w, err)
		return
	}

	log.Info("filter blueprint updated", log.Data{"filter_blueprint_id": filterID, "filter": newFilter})
}

func (api *FilterAPI) getFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	filterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		setErrorCode(w, err)
		return
	}

	bytes, err := json.Marshal(filterOutput)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterOutputID})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	log.Info("got filtered output", log.Data{"filter_output_id": filterOutputID, "filter_output": filterOutput})
}

func (api *FilterAPI) updateFilterOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterOutputID := vars["filter_output_id"]

	if !api.isAuthenticated(r.Header.Get(internalToken)) {
		err := errors.New("Not authorised")
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		setErrorCode(w, err)
		return
	}

	filterOutput, err := models.CreateFilter(r.Body)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterOutputID})
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err := filterOutput.ValidateFilterOutputUpdate(); err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	filterOutput.FilterID = filterOutputID

	// TODO check filter output resource for current downloads
	previousFilterOutput, err := api.dataStore.GetFilterOutput(filterOutputID)
	if err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	updatedFilterOutput := checkFilterOutputQuery(previousFilterOutput, filterOutput)

	if err = api.dataStore.UpdateFilterOutput(updatedFilterOutput); err != nil {
		log.Error(err, log.Data{"filter_output": filterOutput})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("got filtered blueprint", log.Data{"filter_output_id": filterOutputID, "filter_output": filterOutput})
}

func (api *FilterAPI) createFilterOutputResource(newFilter *models.Filter, filterBlueprintID string) models.Filter {
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

	api.dataStore.CreateFilterOutput(&filterOutput)

	api.outputQueue.Queue(&filterOutput)

	return filterOutput
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_blueprint_id"]
	name := vars["name"]
	option := vars["option"]

	err := api.dataStore.RemoveFilterDimensionOption(filterID, name, option)
	if err != nil {
		log.Error(err, log.Data{"filter_blueprint_id": filterID, "dimension": name})
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("delete filtered blueprint", log.Data{"filter_blueprint_id": filterID, "dimension": name})
}

func (api *FilterAPI) getFilterOutputPreview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterID := vars["filter_output_id"]
	requestedLimit := r.URL.Query().Get("limit")
	var limit = 20
	var err error
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Error(errors.New("requested limit is not a number"), log.Data{"filter_output_id": filterID, "limit": limit})
			http.Error(w, "requested limit is not a number", http.StatusBadRequest)
			return
		}
	}
	filterOutput, err := api.dataStore.GetFilterOutput(filterID)
	if err != nil {
		log.ErrorC("failed to find filter output", err, log.Data{"filter_output_id": filterID, "limit": limit})
		setErrorCode(w, err)
		return
	}

	if len(filterOutput.Dimensions) == 0 {
		log.Error(errors.New("no dimensions are present in the filter"), log.Data{"filter_output_id": filterID, "limit": limit})
		http.Error(w, "no dimensions are present in the filter", http.StatusBadRequest)
		return
	}

	data, err := api.preview.GetPreview(filterOutput, limit)
	if err != nil {
		log.ErrorC("failed to query the graph database", err, log.Data{"filter_output_id": filterID, "limit": limit})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error(err, log.Data{"filter_output_id": filterID, "limit": limit})
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"filter_id": filterID, "limit": limit})
		setErrorCode(w, err)
		return
	}

	log.Info("preview filter output", log.Data{"filter_output_id": filterID, "limit": limit, "dimensions": filterOutput.Dimensions})
}

func (api *FilterAPI) isAuthenticated(header string) bool {
	if header == api.internalToken {
		return true
	}

	return false
}

func (api *FilterAPI) checkFilterOptions(ctx context.Context, newFilter *models.Filter, instance *models.Instance) error {
	// Call dimensions list endpoint
	datasetDimensions, err := api.datasetAPI.GetVersionDimensions(ctx, instance.Links.Dataset.ID, instance.Links.Edition.ID, instance.Links.Version.ID)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimensions from the dataset API", err, log.Data{"new_filter": newFilter, "instance": instance})
		return err
	}

	log.Info("dimensions retreived from dataset API", log.Data{"dataset_dimensions": datasetDimensions})

	if err := models.ValidateFilterDimensions(newFilter.Dimensions, datasetDimensions); err != nil {
		log.Error(err, nil)
		return err
	}

	log.Info("successfully validated filter dimensions", log.Data{"filter_dimensions": newFilter.Dimensions})

	var incorrectDimensionOptions []string
	for _, filterDimension := range newFilter.Dimensions {
		// Call dimension options list endpoint
		datasetDimensionOptions, err := api.datasetAPI.GetVersionDimensionOptions(ctx, instance.Links.Dataset.ID, instance.Links.Edition.ID, instance.Links.Version.ID, filterDimension.Name)
		if err != nil {
			log.ErrorC("failed to retreive a list of dimension options from dataset API", err, log.Data{"new_filter": newFilter, "filter_dimension": filterDimension})
			return err
		}

		log.Info("dimension options retreived from dataset API", log.Data{"dimension": filterDimension, "dataset_dimension_option": datasetDimensionOptions})

		incorrectOptions := models.ValidateFilterDimensionOptions(filterDimension.Options, datasetDimensionOptions)
		if incorrectOptions != nil {
			incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
		}
	}

	if incorrectDimensionOptions != nil {
		err = fmt.Errorf("Bad request - incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.ErrorC("Incorrect dimension options chosen", err, log.Data{"dimension_options": incorrectDimensionOptions})
		return err
	}

	return nil
}

func (api *FilterAPI) checkNewFilterDimension(ctx context.Context, newDimension *models.AddDimension, instance *models.Instance) error {
	// FIXME - We should be calling dimension endpoint on dataset API to check if
	// dimension exists but this endpoint doesn't exist yet so call dimension
	// list endpoint and iterate over items to find if dimension exists
	datasetDimensions, err := api.datasetAPI.GetVersionDimensions(ctx, instance.Links.Dataset.ID, instance.Links.Edition.ID, instance.Links.Version.ID)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimensions from the dataset API", err, log.Data{"instance": instance})
		return err
	}

	dimension := models.Dimension{
		Name:    newDimension.Name,
		Options: newDimension.Options,
	}

	if err := models.ValidateFilterDimensions([]models.Dimension{dimension}, datasetDimensions); err != nil {
		log.Error(err, nil)
		return err
	}

	// Call dimension options endpoint
	datasetDimensionOptions, err := api.datasetAPI.GetVersionDimensionOptions(ctx, instance.Links.Dataset.ID, instance.Links.Edition.ID, instance.Links.Version.ID, dimension.Name)
	if err != nil {
		log.ErrorC("failed to retreive a list of dimension options from the dataset API", err, log.Data{"instance": instance})
		return err
	}

	var incorrectDimensionOptions []string
	incorrectOptions := models.ValidateFilterDimensionOptions(dimension.Options, datasetDimensionOptions)
	if incorrectOptions != nil {
		incorrectDimensionOptions = append(incorrectDimensionOptions, incorrectOptions...)
	}

	if incorrectDimensionOptions != nil {
		err = fmt.Errorf("Bad request - incorrect dimension options chosen: %v", incorrectDimensionOptions)
		log.ErrorC("incorrect dimension options chosen", err, log.Data{"dimension_options": incorrectDimensionOptions})
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
	log.Debug("error is", log.Data{"error": err})
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
	case "Instance not found":
		if typ != nil {
			if typ[0] == statusBadRequest {
				http.Error(w, "Bad request - instance not found", http.StatusBadRequest)
				return
			}
			if typ[0] == statusUnprocessableEntity {
				http.Error(w, "Unprocessable entity - instance for filter blueprint no longer exists", http.StatusUnprocessableEntity)
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
	case "Forbidden":
		http.Error(w, forbidden, http.StatusForbidden)
		return
	case "Not authorised":
		http.Error(w, unauthorised, http.StatusUnauthorized)
		return
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
