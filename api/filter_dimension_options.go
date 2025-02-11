package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-filter-api/utils"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-net/v2/links"
	"github.com/ONSdigital/log.go/v2/log"
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
	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	ctx := r.Context()
	log.Info(ctx, "get filter blueprint dimension options", logData)

	offset := api.defaultOffset
	limit := api.defaultLimit
	var err error

	if offsetParameter != "" {
		logData["offset"] = offsetParameter
		offset, err = validatePositiveInt(offsetParameter)
		if err != nil {
			log.Error(ctx, "failed to obtain offset from request query parameters", err, logData)
			setErrorCode(w, err)
			return
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = validatePositiveInt(limitParameter)
		if err != nil {
			log.Error(ctx, "failed to obtain limit from request query parameters", err, logData)
			setErrorCode(w, err)
			return
		}
	}

	if limit > api.maxLimit {
		logData["max_limit"] = api.maxLimit
		err = filters.ErrInvalidQueryParameter
		log.Error(ctx, "limit is greater than the maximum allowed", err, logData)
		setErrorCode(w, err)
		return
	}

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "failed to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	options, err := api.getFilterBlueprintDimensionOptions(ctx, filter, dimensionName, offset, limit)
	if err != nil {
		log.Error(ctx, "failed to get dimension options for filter blueprint", err, logData)
		setErrorCode(w, err)
		return
	}

	// The additions of `options` has been commented out below because sometimes it is resulting
	// in a log line that is greater or equal to: 270836 bytes
	// ... and this is contributing to the 'logstash' servers having a BAD day.
	//logData["options"] = options

	if api.enableURLRewriting {
		dimensionSearchAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)

		for i := range options.Items {
			item := options.Items[i].Links
			//self
			newSelfLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Self.HRef)
			if err == nil {
				options.Items[i].Links.Self.HRef = newSelfLink
			}
			//filter
			newFilterLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Filter.HRef)
			if err == nil {
				options.Items[i].Links.Filter.HRef = newFilterLink
			}
			//Dimension
			newOptionsLink, err := dimensionSearchAPILinksBuilder.BuildLink(item.Dimension.HRef)
			if err == nil {
				options.Items[i].Links.Dimension.HRef = newOptionsLink
			}
		}
	}

	b, err := json.Marshal(options)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimension options into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "got dimension options for filter blueprint", logData)
}

// utility function to cut a slice according to the provided offset and limit.
// Values higher than the slice length are ignored
func slice(full []string, offset, limit int) (sliced []string) {
	end := offset + limit
	if end > len(full) {
		end = len(full)
	}

	if offset > len(full) || limit == 0 {
		return []string{}
	}
	return full[offset:end]
}

func (api *FilterAPI) getFilterBlueprintDimensionOptions(ctx context.Context, filter *models.Filter, dimensionName string, offset, limit int) (options *models.PublicDimensionOptions, err error) {

	for _, dimension := range filter.Dimensions {
		if dimension.Name == dimensionName {

			options = &models.PublicDimensionOptions{
				Items:      []*models.PublicDimensionOption{},
				TotalCount: len(dimension.Options),
				Offset:     offset,
				Limit:      limit,
			}

			// sort alphabetically and cut according to limit and offset
			sort.Strings(dimension.Options)
			dimension.Options = slice(dimension.Options, offset, limit)

			dimLink := fmt.Sprintf("%s/filters/%s/dimensions/%s", api.host, filter.FilterID, dimension.Name)
			filterObject := &models.LinkObject{
				HRef: fmt.Sprintf("%s/filters/%s", api.host, filter.FilterID),
				ID:   filter.FilterID,
			}

			for _, option := range dimension.Options {
				dimensionOption := &models.PublicDimensionOption{
					Links: &models.PublicDimensionOptionLinkMap{
						Self:      &models.LinkObject{HRef: dimLink + "/options/" + option, ID: option},
						Dimension: &models.LinkObject{HRef: dimLink, ID: dimension.Name},
						Filter:    filterObject,
					},
					Option: option,
				}
				options.Items = append(options.Items, dimensionOption)
			}
			options.Count = len(options.Items)
			return options, nil
		}
	}

	return nil, filters.ErrDimensionNotFound
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
	log.Info(ctx, "get filter blueprint dimension option", logData)

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID, mongo.AnyETag)
	if err != nil {
		log.Error(ctx, "unable to get dimension option for filter blueprint", err, logData)
		setErrorCodeFromError(w, err)
		return
	}

	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filter, dimensionName, option)
	if err != nil {
		log.Error(ctx, "unable to get dimension option for filter blueprint", err, logData)
		setErrorCodeFromError(w, err)
		return
	}

	if api.enableURLRewriting {
		dimensionSearchAPILinksBuilder := links.FromHeadersOrDefault(&r.Header, api.host)

		linkFields := map[string]*models.LinkObject{
			"Self":      dimensionOption.Links.Self,
			"Filter":    dimensionOption.Links.Filter,
			"Dimension": dimensionOption.Links.Dimension,
		}

		for _, linkObj := range linkFields {
			if linkObj != nil && linkObj.HRef != "" {
				newLink, err := dimensionSearchAPILinksBuilder.BuildLink(linkObj.HRef)
				if err == nil {
					linkObj.HRef = newLink
				} else {
					log.Error(ctx, "failed to rewrite dimension option link", err, logData)
					setErrorCode(w, err)
				}
			}
		}
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimension option into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, filter.ETag)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "got dimension option for filter blueprint", logData)
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
					filterObject := &models.LinkObject{
						HRef: fmt.Sprintf("%s/filters/%s", api.host, filter.FilterID),
						ID:   filter.FilterID,
					}

					dimensionOption = &models.PublicDimensionOption{
						Links: &models.PublicDimensionOptionLinkMap{
							Self:      &models.LinkObject{HRef: dimLink + "/options/" + option, ID: option},
							Dimension: &models.LinkObject{HRef: dimLink, ID: d.Name},
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

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// add the dimension options, if valid
	newETag, err := api.addFilterBlueprintDimensionOptions(ctx, filterBlueprintID, dimensionName, []string{option}, logData, eTag)
	if err != nil {
		log.Error(ctx, "error adding filter blueprint dimension option", err, logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	// request filterBlueprint again in order to construct the response from the updated filter (if a new option was added)
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID, newETag)
	if err != nil {
		log.Error(ctx, "error getting filter blueprint dimension option after the dimension option has been successfully added", err, logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	// get the options from the new filterBlueprint
	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filterBlueprint, dimensionName, option)
	if err != nil {
		log.Error(ctx, "unable to get dimension option for filter blueprint", err, logData)
		setErrorCodeFromErrorExpectDimension(w, err)
		return
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Error(ctx, "failed to marshal filter blueprint dimension option into bytes", err, logData)
		http.Error(w, InternalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	setETag(w, newETag)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err, logData)
		setErrorCode(w, err)
		return
	}

	log.Info(ctx, "created new dimension option for filter blueprint", logData)
}

// addFilterBlueprintDimensionOptions adds the provided options to the filter dimension, only if the options are available for the dimension.
func (api *FilterAPI) addFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprintID string, dimensionName string, options []string, logData log.Data, eTag string) (newETag string, err error) {

	// request filterBlueprint before updating it
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID, eTag)
	if err != nil {
		return "", err
	}

	// Check if dimension exists and any provided option already exists
	hasDimension, _, missingOptions := findDimensionAndOptions(filterBlueprint, dimensionName, options)
	if !hasDimension {
		return "", filters.ErrDimensionNotFound
	}

	// validate that the provided existing dimension is still valid and the options are acceptable for the dimension
	if err := api.checkNewFilterDimension(ctx, dimensionName, options, filterBlueprint.Dataset); err != nil {
		if err == filters.ErrVersionNotFound || err == filters.ErrDimensionsNotFound {
			return "", err
		}

		if incorrectDimensionOptions.MatchString(err.Error()) {
			return "", filters.NewBadRequestErr(err.Error())
		}

		if incorrectDimension.MatchString(err.Error()) {
			return "", filters.NewBadRequestErr(err.Error())
		}

		return "", err
	}

	// All validations succeeded - add dimension options that do not already exist
	return api.dataStore.AddFilterDimensionOptions(ctx, filterBlueprintID, dimensionName, utils.CreateArray(missingOptions), filterBlueprint.UniqueTimestamp, filterBlueprint.ETag, filterBlueprint)
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
	log.Info(ctx, "remove filter blueprint dimension option", logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newETag, err := api.removeFilterBlueprintDimensionOption(ctx, filterBlueprintID, dimensionName, option, eTag)
	if err != nil {
		log.Error(ctx, "error removing filter blueprint dimension option", err, logData)
		setErrorCodeFromError(w, err)
		return
	}

	setJSONContentType(w)
	setETag(w, newETag)
	w.WriteHeader(http.StatusNoContent)

	log.Info(ctx, "delete dimension option on filter blueprint", logData)
}

// removeFilterBlueprintDimensionOption removes a single dimension option, failing if the option did not exist
func (api *FilterAPI) removeFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID string, dimensionName, option, eTag string) (newETag string, err error) {

	// Get filter Blueprint before updating it
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID, eTag)
	if err != nil {
		return "", err
	}

	// Check if dimension and option exists
	hasDimension, hasOptions, _ := findDimensionAndOptions(filterBlueprint, dimensionName, []string{option})

	if !hasDimension {
		return "", filters.ErrDimensionNotFound
	}

	if !hasOptions {
		return "", filters.ErrDimensionOptionNotFound
	}

	return api.dataStore.RemoveFilterDimensionOption(ctx, filterBlueprint.FilterID, dimensionName, option, filterBlueprint.UniqueTimestamp, filterBlueprint.ETag, filterBlueprint)
}

// removeFilterBlueprintDimensionOption removes all provided options.
func (api *FilterAPI) removeFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprintID string, dimensionName string, options []string, logData log.Data, eTag string) (newETag string, err error) {

	// check if any option has been provided
	if len(options) == 0 {
		return eTag, nil
	}

	// Get filter Blueprint before updating it
	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID, eTag)
	if err != nil {
		return "", err
	}

	// Check if provided dimension and options exists in filter blueprint
	hasDimension, hasAllOptions, missingOptions := findDimensionAndOptions(filterBlueprint, dimensionName, options)
	if !hasDimension {
		return "", filters.ErrDimensionNotFound
	}

	// find options that actually need to be removed according to the existing options before applying any change
	optionsToRemove := []string{}
	if hasAllOptions {
		optionsToRemove = options
	} else {
		for _, option := range options {
			if _, found := missingOptions[option]; !found {
				optionsToRemove = append(optionsToRemove, option)
			}
		}
	}

	// if none of the provided options were present, we don't need to remove anything
	if len(optionsToRemove) == 0 {
		log.Info(ctx, "options do not exist in the dimension, nothing to remove")
		return eTag, nil
	}

	// remove necessary options from DB
	return api.dataStore.RemoveFilterDimensionOptions(ctx, filterBlueprintID, dimensionName, optionsToRemove, filterBlueprint.UniqueTimestamp, eTag, filterBlueprint)
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
	log.Info(ctx, "patch filter blueprint dimension", logData)

	// eTag value must be present in If-Match header
	eTag, err := getIfMatchForce(r)
	if err != nil {
		log.Error(ctx, "missing header", err, log.Data{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// unmarshal and validate the patch array
	patches, err := models.CreatePatches(r.Body)
	if err != nil {
		log.Error(ctx, "error obtaining patch from request body", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logData["patch_list"] = patches

	// check that the provided paths are acceptable and the total values do not exceed the maximum allowed
	totalValues := 0
	for _, patch := range patches {
		if patch.Path != "/options/-" {
			err = fmt.Errorf("provided path '%s' not supported. Supported paths: '/options/-'", patch.Path)
			log.Error(ctx, "error validating patch operation path, no change has been applied", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		v2, err := getStringArrayFromInterface(patch.Value)
		if err != nil {
			err = fmt.Errorf("values provided are not strings")
			log.Error(ctx, "error validating patch operation path, no change has been applied", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		totalValues += len(v2)
		if totalValues > api.maxRequestOptions {
			logData["max_options"] = api.maxRequestOptions
			err = fmt.Errorf("a maximum of %d overall option values can be provied in a set of patch operations, which has been exceeded", api.maxRequestOptions)
			log.Error(ctx, "error validating patch operation values size, no change has been applied", err, logData)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// apply the patches to the filter blueprint dimension options
	var newETag interface{}
	newETag, err = api.dataStore.RunTransaction(ctx, true, func(txCtx context.Context) (interface{}, error) {
		var (
			allOptions, options []string
			tag                 = eTag
		)

		// apply patch operations sequentially, stop processing if one patch fails, and return a list of successful patches operations
		for _, patch := range patches {
			allOptions, err = getStringArrayFromInterface(patch.Value)
			if err != nil {
				return tag, err
			}
			options = RemoveDuplicateAndEmptyOptions(allOptions)

			if patch.Op == dprequest.OpAdd.String() {
				tag, err = api.addFilterBlueprintDimensionOptions(txCtx, filterBlueprintID, dimensionName, options, logData, tag)
				if err != nil {
					return tag, err
				}
			} else {
				tag, err = api.removeFilterBlueprintDimensionOptions(txCtx, filterBlueprintID, dimensionName, options, logData, tag)
				if err != nil {
					return tag, err
				}
			}
		}

		return tag, nil
	})
	if err != nil {
		log.Error(ctx, "error patching filter blueprint dimension options", err, logData)
		setErrorCodeFromError(w, err)
		return
	}
	t, ok := newETag.(string)
	if !ok {
		err = errors.New("returned ETag is not a string value")
		log.Error(ctx, err.Error(), err, log.Data{"newETag": newETag})
		setErrorCodeFromError(w, err)
		return
	}
	setETag(w, t)

	// set content type, marshal and write response
	setJSONPatchContentType(w)
	if err = WriteJSONBody(ctx, patches, w, logData); err != nil {
		log.Error(ctx, "error writing JSON body after a successful filter blueprint patch", err, logData)
		setErrorCodeFromError(w, err)
		return
	}

	log.Info(ctx, "successfully patched filter dimension options on filter blueprint", logData)
}

// findDimensionAndOptions finds the provided dimensionName and options (in the dimension) in the filterBlueprint
func findDimensionAndOptions(filterBlueprint *models.Filter, dimensionName string, options []string) (hasDimension bool, hasAllOptions bool, missingOptions map[string]struct{}) {

	// unique option names that have not been found yet
	missingOptions = utils.CreateMap(options)

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
	case filters.ErrFilterBlueprintNotFound, filters.ErrInvalidQueryParameter, filters.ErrDimensionNotFound:
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
