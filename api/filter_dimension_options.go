package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"

	"context"
	"fmt"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/go-ns/common"
)

const (
	// audit actions
	getOptionsAction   = "getFilterBlueprintDimensionOptions"
	getOptionAction    = "getFilterBlueprintDimensionOption"
	removeOptionAction = "removeFilterBlueprintDimensionOption"
	addOptionAction    = "addFilterBlueprintDimensionOption"
)

func (api *FilterAPI) getFilterBlueprintDimensionOptionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filterBlueprintID := vars["filter_blueprint_id"]
	dimensionName := vars["name"]
	logData := log.Data{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"action":              getOptionsAction,
	}

	ctx := r.Context()
	log.Event(ctx, "get filter blueprint dimension options", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
	}
	if auditErr := api.auditor.Record(ctx, getOptionsAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getOptionsAction, actionAttempted, w, auditErr, logData)
		return
	}

	options, err := api.getFilterBlueprintDimensionOptions(ctx, filterBlueprintID, dimensionName)
	if err != nil {
		log.Event(ctx, "failed to get dimension options for filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		setErrorCode(w, err)
		return
	}

	logData["options"] = options

	b, err := json.Marshal(options)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension options into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionsAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionsAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, getOptionsAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getOptionsAction, actionSuccessful, w, auditErr, logData)
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

func (api *FilterAPI) getFilterBlueprintDimensionOptions(ctx context.Context, filterBlueprintID, dimensionName string) ([]models.PublicDimensionOption, error) {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return nil, err
	}

	var options []models.PublicDimensionOption
	dimensionFound := false
	for _, dimension := range filter.Dimensions {

		if dimension.Name == dimensionName {
			dimensionFound = true

			dimLink := fmt.Sprintf("%s/filters/%s/dimensions/%s", api.host, filterBlueprintID, dimension.Name)
			filterObject := models.LinkObject{
				HRef: fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID),
				ID:   filterBlueprintID,
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
		"action":              getOptionAction,
	}

	ctx := r.Context()
	log.Event(ctx, "get filter blueprint dimension option", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	if auditErr := api.auditor.Record(ctx, getOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Event(ctx, "unable to get dimension option for filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		default:
			setErrorCode(w, err)
		}
		return
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension option into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, getOptionAction, actionSuccessful, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, getOptionAction, actionSuccessful, w, auditErr, logData)
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

func (api *FilterAPI) getFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) (*models.PublicDimensionOption, error) {

	filter, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return nil, err
	}

	optionFound := false
	dimensionFound := false

	var dimensionOption *models.PublicDimensionOption

	for _, d := range filter.Dimensions {
		if d.Name == dimensionName {
			dimensionFound = true
			for _, o := range d.Options {

				if o == option {
					optionFound = true

					dimLink := fmt.Sprintf("%s/filters/%s/dimensions/%s", api.host, filterBlueprintID, d.Name)
					filterObject := models.LinkObject{
						HRef: fmt.Sprintf("%s/filters/%s", api.host, filterBlueprintID),
						ID:   filterBlueprintID,
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
		"action":              addOptionAction,
	}

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	ctx := r.Context()
	if auditErr := api.auditor.Record(ctx, addOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, addOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	err := api.addFilterBlueprintDimensionOption(ctx, filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Event(ctx, "error adding filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, addOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, addOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
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
		return
	}

	err = api.checkDisclosureStatus(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "disclosure control update failure", log.Error(err), log.ERROR)
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
		return
	}

	dimensionOption, err := api.getFilterBlueprintDimensionOption(ctx, filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Event(ctx, "unable to get dimension option for filter blueprint", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		default:
			setErrorCode(w, err)
		}
		return
	}

	b, err := json.Marshal(dimensionOption)
	if err != nil {
		log.Event(ctx, "failed to marshal filter blueprint dimension option into bytes", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, getOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, getOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if auditErr := api.auditor.Record(ctx, addOptionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(ctx, addOptionAction, actionSuccessful, auditErr, logData)
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

func (api *FilterAPI) addFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) error {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	timestamp := filterBlueprint.UniqueTimestamp

	// FIXME - Once dataset API has an endpoint to check single option exists,
	// refactor code below instead of creating an AddDimension object from the
	// AddDimensionOption object (to be able to use checkNewFilterDimension method)
	if err = api.checkNewFilterDimension(ctx, dimensionName, []string{option}, filterBlueprint.Dataset); err != nil {
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

	if err := api.dataStore.AddFilterDimensionOption(filterBlueprintID, dimensionName, option, timestamp); err != nil {
		return err
	}

	return nil
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
		"action":              removeOptionAction,
	}
	ctx := r.Context()
	log.Event(ctx, "remove filter blueprint dimension option", log.INFO, logData)

	auditParams := common.Params{
		"filter_blueprint_id": filterBlueprintID,
		"dimension":           dimensionName,
		"option":              option,
	}
	if auditErr := api.auditor.Record(ctx, removeOptionAction, actionAttempted, auditParams); auditErr != nil {
		handleAuditingFailure(ctx, removeOptionAction, actionAttempted, w, auditErr, logData)
		return
	}

	err := api.removeFilterBlueprintDimensionOption(ctx, filterBlueprintID, dimensionName, option)
	if err != nil {
		log.Event(ctx, "error removing filter blueprint dimension option", log.ERROR, log.Error(err), logData)
		if auditErr := api.auditor.Record(ctx, removeOptionAction, actionUnsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(ctx, removeOptionAction, actionUnsuccessful, w, auditErr, logData)
			return
		}
		switch err {
		case filters.ErrFilterBlueprintNotFound:
			setErrorCode(w, err, statusBadRequest)
		case filters.ErrVersionNotFound:
			setErrorCode(w, err, statusUnprocessableEntity)
		default:
			setErrorCode(w, err)
		}
		return
	}

	err = api.checkDisclosureStatus(ctx, filterBlueprintID)
	if err != nil {
		log.Event(ctx, "disclosure control update failure", log.Error(err), log.ERROR)
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
		return
	}

	if auditErr := api.auditor.Record(ctx, removeOptionAction, actionSuccessful, auditParams); auditErr != nil {
		logAuditFailure(ctx, removeOptionAction, actionSuccessful, auditErr, logData)
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusNoContent)

	log.Event(ctx, "delete dimension option on filter blueprint", log.INFO, logData)
}

func (api *FilterAPI) removeFilterBlueprintDimensionOption(ctx context.Context, filterBlueprintID, dimensionName, option string) error {

	filterBlueprint, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	timestamp := filterBlueprint.UniqueTimestamp

	// Check if dimension and option exists
	var hasDimension bool
	var hasOption bool
	for _, dimension := range filterBlueprint.Dimensions {
		if dimension.Name == dimensionName {
			hasDimension = true
			for _, dimOption := range dimension.Options {
				if dimOption == option {
					hasOption = true
					break
				}
			}
			break
		}
	}

	if !hasDimension {
		return filters.ErrDimensionNotFound
	}

	if !hasOption {
		return filters.ErrDimensionOptionNotFound
	}

	if err = api.dataStore.RemoveFilterDimensionOption(filterBlueprintID, dimensionName, option, timestamp); err != nil {
		return err
	}

	return nil
}

func (api *FilterAPI) checkDisclosureStatus(ctx context.Context, filterBlueprintID string) error {
	f, err := api.getFilterBlueprint(ctx, filterBlueprintID)
	if err != nil {
		return err
	}

	result, err := api.doFTBQuery(ctx, f)
	if err != nil {
		return err
	}

	err = api.updateFilterDisclosureStatus(f, result)
	if err != nil {
		return err
	}
	return nil
}

func (api *FilterAPI) doFTBQuery(ctx context.Context, blueprint *models.Filter) (*ftb.QueryResult, error) {
	cli := ftb.NewClient("http://localhost:10100", os.Getenv("AUTH_PROXY_TOKEN"), dphttp.DefaultClient)

	dimensions := make([]ftb.DimensionOptions, 0)
	for _, d := range blueprint.Dimensions {
		dimensions = append(dimensions, ftb.DimensionOptions{
			Name:    d.Name,
			Options: d.Options,
		})
	}

	query := ftb.Query{
		RootDimension:     "OA",
		DatasetName:       "People",
		DimensionsOptions: dimensions,
	}

	return cli.Query(ctx, query)
}

func (api *FilterAPI) updateFilterDisclosureStatus(filter *models.Filter, result *ftb.QueryResult) error {
	status := result.Status
	dimension := ""
	blockedOptions := make([]string, 0)

	if result.IsBlockedByRules() {
		dimension = result.DisclosureControlDetails.Dimension
		blockedOptions = result.DisclosureControlDetails.BlockedOptions
	}

	return api.dataStore.UpdateDisclosureStatus(filter.FilterID, status, dimension, blockedOptions, filter.UniqueTimestamp)
}
