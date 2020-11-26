package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/models"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"

	"context"
	"fmt"

	"github.com/ONSdigital/dp-filter-api/filters"
)

func (api *FilterAPI) getFilterBlueprintDimensionOptionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}

	ctx := r.Context()
	log.Event(ctx, "get filter blueprint dimension options", log.INFO, logData)

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "failed to get dimension options for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	options, err := api.getFilterBlueprintDimensionOptions(ctx, filter, dimensionName)
	if err != nil {
		log.Event(ctx, "failed to get dimension options for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	logData["options"] = options

	b, err := json.Marshal(options)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension options into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "got dimension options for filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(ctx context.Context, filter *models.Filter, dimensionName string) ([]models.PublicDimensionOption, error) {

	var options []models.PublicDimensionOption
	dimensionFound := false
	for _, dimension := range filter.Dimensions {

		if dimension.Name == dimensionName {
			dimensionFound = true

			dimLink := fmt.Sprintf("%s/filters/%s/dimensions/%s", api.host, filter.FilterID, dimension.Name)
			filterObject := models.LinkObject{
				HRef: fmt.Sprintf("%s/filters/%s", api.host, filter.FilterID),
				ID:   filter.FilterID,
			}

			for _, option := range dimension.Options {
				dimensionOption := &models.PublicDimensionOption{
					Links: &models.PublicDimensionOptionLinkMap{
						Self:      models.LinkObject{HRef: dimLink + "/options/" + option, ID: option},
						Dimension: models.LinkObject{HRef: dimLink, ID: dimension.Name},
						Filter:    filterObject,
					},
					Option: option,
				}
				options = append(options, *dimensionOption)
			}
		}
	}

	if !dimensionFound {
		return nil, filters.ErrDimensionNotFound
	}

	return options, nil
}

func (api *FilterAPI) getFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}

	ctx := r.Context()
	log.Event(ctx, "get filter blueprint dimension option", log.INFO, logData)

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "unable to get dimension option for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filter, dimensionName, option)
	if err != nil {
		log.Event(ctx, "unable to get dimension option for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension option into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "got dimension option for filter blueprint", log.INFO, logData)
}

// obtain the dimension from the provided filter model with name == dimensionName
func (api *FilterAPI) getDimension(ctx context.Context, filter *models.Filter, dimensionName string) (*models.Dimension, error) {
	for _, d := range filter.Dimensions {
		if d.Name == dimensionName {
			return &d, nil
		}
	}
	return nil, filters.ErrDimensionNotFound
}

func (api *FilterAPI) getFilterBlueprintDimensionOption(ctx context.Context, filter *models.Filter, dimensionName, option string) (*models.PublicDimensionOption, error) {

	optionFound := false
	dimensionFound := false

	var dimensionOption *models.PublicDimensionOption

	for _, d := range filter.Dimensions {
		if d.Name == dimensionName {
			dimensionFound = true
			for _, o := range d.Options {

				if o == option {
					optionFound = true

					dimLink := fmt.Sprintf("%s/filters/%s/dimensions/%s", api.host, filter.FilterID, d.Name)
					filterObject := models.LinkObject{
						HRef: fmt.Sprintf("%s/filters/%s", api.host, filter.FilterID),
						ID:   filter.FilterID,
					}

					dimensionOption = &models.PublicDimensionOption{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      models.LinkObject{HRef: dimLink + "/options/" + option, ID: option},
							Dimension: models.LinkObject{HRef: dimLink, ID: d.Name},
							Filter:    filterObject,
						},
						Option: option,
					}
					break
				}
			}
			break
		}
	}

	if !dimensionFound {
		return nil, filters.ErrDimensionNotFound
	}

	if !optionFound {
		return nil, filters.ErrDimensionOptionNotFound
	}

	return dimensionOption, nil
}

func (api *FilterAPI) addFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension_name":      dimensionName,
		"dimension_option":    option,
	}
	ctx := r.Context()

	// request filterBlueprint before updating it
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "error getting filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	// add the dimension options if not presentan dimension and options are valid
	err = api.addFilterBlueprintDimensionOptions(ctx, filterBlueprint, dimensionName, []string{option}, logData)
	if err != nil {
		log.Event(ctx, "error adding filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	// request filterBlueprint again in order to construct the response from the updated filter (if a new option was added)
	filterBlueprint, err = api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "error getting filter blueprint dimension option after the dimension option has been successfully added", log.ERROR, log.Error(err), logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	// get the options from the new filterBlueprint
	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filterBlueprint, dimensionName, option)
	if err != nil {
		log.Event(ctx, "unable to get dimension option for filter blueprint", log.ERROR, log.Error(err), logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension option into bytes", log.ERROR, log.Error(err), logData)
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		log.Event(ctx, "failed to write bytes for http response", log.ERROR, log.Error(err), logData)
		setErrorCode(w, err)
		return
	}

	log.Event(ctx, "created new dimension option for filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) addFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprint *models.Filter, dimensionName string, options []string, logData log.Data) error {

	timestamp := filterBlueprint.UniqueTimestamp

	// Check if dimension exists and any provided option already exists
	hasDimension, _, missingOptions := findDimensionAndOptions(filterBlueprint, dimensionName, options)
	if !hasDimension {
		return filters.ErrDimensionNotFound
	}

	// validate that the provided existing dimension is still valid and the options are acceptable for the dimension
	if err := api.checkNewFilterDimension(ctx, dimensionName, options, filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			return err
		}

		if incorrectDimensionOptions.MatchString(err.Error()) {
			return filters.NewBadRequestErr(err.Error())
		}

		if incorrectDimension.MatchString(err.Error()) {
			return filters.NewBadRequestErr(err.Error())
		}

		return err
	}

	// All validations succeeded - add dimension options that do not already exist
	return api.dataStore.AddFilterDimensionOptions(filterBlueprint.FilterID, dimensionName, createArray(missingOptions), timestamp)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	option := vars["option"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	ctx := r.Context()
	log.Event(ctx, "remove filter blueprint dimension option", log.INFO, logData)

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "error removing filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	err = api.removeFilterBlueprintDimensionOption(ctx, filterBlueprint, dimensionName, option)
	if err != nil {
		log.Event(ctx, "error removing filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Event(ctx, "delete dimension option on filter blueprint", log.INFO, logData)
}

// removeFilterBlueprintDimensionOption removes a single dimension option, failing if the option did not exist
func (api *FilterAPI) removeFilterBlueprintDimensionOption(ctx context.Context, filterBlueprint *models.Filter, dimensionName, option string) error {

	// Check if dimension and option exists
	hasDimension, hasOptions, _ := findDimensionAndOptions(filterBlueprint, dimensionName, []string{option})

	if !hasDimension {
		return filters.ErrDimensionNotFound
	}

	if !hasOptions {
		return filters.ErrDimensionOptionNotFound
	}

	if err := api.dataStore.RemoveFilterDimensionOption(filterBlueprint.FilterID, dimensionName, option, filterBlueprint.UniqueTimestamp); err != nil {
		return err
	}

	return nil
}

// removeFilterBlueprintDimensionOption removes all provided options, if any option did not exist, it will be ignored
func (api *FilterAPI) removeFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprint *models.Filter, dimensionName string, options []string, logData log.Data) error {

	// check if any option has been provided
	if len(options) == 0 {
		return nil
	}

	// Check if provided dimension and options exists
	hasDimension, hasOptions, missingOptions := findDimensionAndOptions(filterBlueprint, dimensionName, options)

	if !hasDimension {
		return filters.ErrDimensionNotFound
	}

	if !hasOptions {
		log.Event(ctx, "options do not exist in the dimension, nothing to remove", log.INFO, log.Data{})
		return nil
	}

	timestamp := filterBlueprint.UniqueTimestamp

	// find options that need to be removed (i.e. are present in filer blueprint)
	optionsToRemove := []string{}
	for _, option := range options {
		if _, found := missingOptions[option]; !found {
			optionsToRemove = append(optionsToRemove, option)
		}
	}

	// remove necessary options from DB
	return api.dataStore.RemoveFilterDimensionOptions(filterBlueprint.FilterID, dimensionName, optionsToRemove, timestamp)
}

// Handler for a list of patch operations against the dimension options
func (api *FilterAPI) patchFilterBlueprintDimensionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]

	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	ctx := r.Context()
	log.Event(ctx, "patch filter blueprint dimension", log.INFO, logData)

	// unmarshal and validate the patch array
	patches, err := models.CreatePatches(r.Body)
	if err != nil {
		log.Event(ctx, "error obtaining patch from request body", log.ERROR, log.Error(err), logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logData["patch_list"] = patches

	// check that the provided paths are acceptable and the total values do not exceed the maximum allowed
	totalValues := 0
	for _, patch := range patches {
		if patch.Path != "/options/-" {
			err = fmt.Errorf("provided path '%s' not supported. Supported paths: '/options/-'", patch.Path)
			log.Event(ctx, "error validating patch operation path, no change has been applied", log.ERROR, log.Error(err), logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		totalValues += len(patch.Value)
		if totalValues > api.maxRequestOptions {
			logData["max_options"] = api.maxRequestOptions
			err = fmt.Errorf("a maximum of %d overall option values can be provied in a set of patch operations, which has been exceeded", api.maxRequestOptions)
			log.Event(ctx, "error validating patch operation values size, no change has been applied", log.ERROR, log.Error(err), logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Get filter blueprint
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "error patching filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	// apply the patches to the filter blueprint dimension options
	successfulPatches, err := api.patchFilterBlueprintDimension(ctx, filterBlueprint, dimensionName, patches, logData)
	if err != nil {
		logData["successful_patches"] = successfulPatches
		log.Event(ctx, "error patching filter blueprint dimension options", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		if len(successfulPatches) > 0 {
			if err := WriteJSONBody(ctx, successfulPatches, w, logData); err != nil {
				log.Event(ctx, "error writing JSON body during filter blueprint patch error handling", log.ERROR, log.Error(err), logData)
			}
		}
		return
	}

	// set content type, marshal and write response
	setJSONPatchContentType(w)
	if err := WriteJSONBody(ctx, successfulPatches, w, logData); err != nil {
		log.Event(ctx, "error writing JSON body after a successful filter blueprint patch", log.ERROR, log.Error(err), logData)
		setErrorCodeFromError(w, err)
		return
	}

	log.Event(ctx, "successfully patched filter dimension options on filter blueprint", log.INFO, logData)
}

// patchFilterBlueprintDimension applies the patches by calling add or remove filter dimension options. It keeps track of a list of successful patches, so that
func (api *FilterAPI) patchFilterBlueprintDimension(ctx context.Context, filterBlueprint *models.Filter, dimensionName string, patches []dprequest.Patch, logData log.Data) (successful []dprequest.Patch, err error) {

	successful = []dprequest.Patch{}

	// if filter blueprint does not have dimension, return now with ErrDimensionNotFound, as no path operation can succeed
	hasDimension, _, _ := findDimensionAndOptions(filterBlueprint, dimensionName, []string{})
	if !hasDimension {
		return successful, filters.ErrDimensionNotFound
	}

	// apply patch operations sequentially, stop processing if one patch fails, and return a list of successful patches operations
	for _, patch := range patches {
		options := removeDuplicateAndEmptyOptions(patch.Value)
		if patch.Op == dprequest.OpAdd.String() {
			if err := api.addFilterBlueprintDimensionOptions(ctx, filterBlueprint, dimensionName, options, logData); err != nil {
				return successful, err
			}
		} else {
			if err := api.removeFilterBlueprintDimensionOptions(ctx, filterBlueprint, dimensionName, options, logData); err != nil {
				return successful, err
			}
		}
		successful = append(successful, patch)
	}
	return successful, nil
}

// findDimensionAndOptions finds the provided dimensionName and options (in the dimension) in the filterBlueprint
func findDimensionAndOptions(filterBlueprint *models.Filter, dimensionName string, options []string) (hasDimension bool, hasAllOptions bool, missingOptions map[string]struct{}) {

	// unique option names that have not been found yet
	missingOptions = createMap(options)

	// find dimension and options in dimension
	for _, dimension := range filterBlueprint.Dimensions {
		if dimension.Name == dimensionName {
			for _, dimOption := range dimension.Options {
				if _, foundOption := missingOptions[dimOption]; foundOption {
					delete(missingOptions, dimOption)
					if len(missingOptions) == 0 {
						return true, true, missingOptions
					}
				}
			}
			return true, false, missingOptions
		}
	}
	return false, false, missingOptions
}

// createArray creates an array of keys from the provided map
func createArray(m map[string]struct{}) (a []string) {
	for k := range m {
		a = append(a, k)
	}
	return a
}

// createMap creates a map whose keys are the unique values of the provided array(s).
// values are empty structs for memory efficiency reasons (no storage used)
func createMap(a ...[]string) (m map[string]struct{}) {
	m = make(map[string]struct{})
	for _, aa := range a {
		for _, val := range aa {
			m[val] = struct{}{}
		}
	}
	return m
}

// setErrorCodeFromError sets the HTTP Status Code according to the provided error.
func setErrorCodeFromError(w http.ResponseWriter, err error) {
	switch err {
	case filters.ErrFilterBlueprintNotFound:
		setErrorCode(w, err, statusBadRequest)
	case filters.ErrDimensionsNotFound:
		fallthrough
	case filters.ErrVersionNotFound:
		setErrorCode(w, err, statusUnprocessableEntity)
	default:
		setErrorCode(w, err)
	}
}

// setErrorCodeFromError sets the HTTP Status Code according to the provided error, expecting the dimension (ErrDimensionNotFound will be mapped to statusBadRequest)
func setErrorCodeFromErrorExpectDimension(w http.ResponseWriter, err error) {
	switch err {
	case filters.ErrFilterBlueprintNotFound:
		setErrorCode(w, err, statusBadRequest)
	case filters.ErrDimensionNotFound:
		setErrorCode(w, err, statusBadRequest)
	case filters.ErrDimensionsNotFound:
		fallthrough
	case filters.ErrVersionNotFound:
		setErrorCode(w, err, statusUnprocessableEntity)
	default:
		setErrorCode(w, err)
	}
}

// WriteJSONBody marshals the provided interface into json, and writes it to the response body.
func WriteJSONBody(ctx context.Context, v interface{}, w http.ResponseWriter, data log.Data) error {

	// Marshal provided model
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Write payload to body
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return nil
}
