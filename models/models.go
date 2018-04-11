package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

// A list of states
const (
	CreatedState   = "created"
	SubmittedState = "submitted"
	CompletedState = "completed"
)

var (
	Unpublished = false
	Published   = true
)

// Dataset contains the uniique identifiers that make a dataset unique
type Dataset struct {
	ID      string `bson:"id"        json:"id"`
	Edition string `bson:"edition"   json:"edition"`
	Version int    `bson:"version"   json:"version"`
}

// NewFilter creates a filter by using the dataset id, edition and version
type NewFilter struct {
	Dataset    *Dataset    `bson:"dataset" json:"dataset"`
	Dimensions []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
}

// Filter represents a structure for a filter job
type Filter struct {
	Dataset     *Dataset    `bson:"dataset"              json:"dataset"`
	InstanceID  string      `bson:"instance_id"          json:"instance_id"`
	Dimensions  []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Downloads   *Downloads  `bson:"downloads,omitempty"  json:"downloads,omitempty"`
	Events      Events      `bson:"events,omitempty"     json:"events,omitempty"`
	FilterID    string      `bson:"filter_id"            json:"filter_id,omitempty"`
	State       string      `bson:"state,omitempty"      json:"state,omitempty"`
	Published   *bool       `bson:"published"            json:"published"`
	Links       LinkMap     `bson:"links"                json:"links,omitempty"`
	LastUpdated time.Time   `bson:"last_updated"         json:"-"`
}

// LinkMap contains a named LinkObject for each link to other resources
type LinkMap struct {
	Dimensions      LinkObject `bson:"dimensions"                 json:"dimensions,omitempty"`
	FilterOutput    LinkObject `json:"filter_output,omitempty"`
	FilterBlueprint LinkObject `bson:"filter_blueprint,omitempty" json:"filter_blueprint,omitempty"`
	Self            LinkObject `bson:"self"                       json:"self,omitempty"`
	Version         LinkObject `bson:"version"                    json:"version,omitempty"`
}

// LinkObject represents a generic structure for all links
type LinkObject struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty"`
	HRef string `bson:"href"         json:"href,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	URL     string   `bson:"dimension_url,omitempty" json:"dimension_url,omitempty"`
	Name    string   `bson:"name"                    json:"name"`
	Options []string `bson:"options"                 json:"options"`
}

// Downloads represents a list of file types possible to download
type Downloads struct {
	CSV *DownloadItem `bson:"csv,omitempty"  json:"csv,omitempty"`
	XLS *DownloadItem `bson:"xls,omitempty"  json:"xls,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	HRef    string `bson:"href,omitempty"    json:"href,omitempty"`
	Private string `bson:"private,omitempty" json:"private,omitempty"`
	Public  string `bson:"public,omitempty"  json:"public,omitempty"`
	Size    string `bson:"size,omitempty"    json:"size,omitempty"`
}

// Events represents a list of array objects containing event information against the filter job
type Events struct {
	Error []EventItem `bson:"error,omitempty" json:"error,omitempty"`
	Info  []EventItem `bson:"info,omitempty"  json:"info,omitempty"`
}

// EventItem represents an event object containing event information
type EventItem struct {
	Message string `bson:"message" json:"message,omitempty"`
	Time    string `bson:"time"    json:"time,omitempty"`
	Type    string `bson:"type"    json:"type,omitempty"`
}

// DimensionOption represents dimension option information
type DimensionOption struct {
	DimensionOptionURL string `json:"dimension_option_url"`
	Option             string `json:"option"`
}

// A list of errors returned from package
var (
	ErrorReadingBody = errors.New("Failed to read message body")
	ErrorParsingBody = errors.New("Failed to parse json body")
	ErrorNoData      = errors.New("Bad request - Missing data in body")
)

// ValidateNewFilter checks the content of the filter structure
func (filter *NewFilter) ValidateNewFilter() error {
	// FilterBluePrint should not have state
	var missingFields []string

	if filter.Dataset != nil {
		if filter.Dataset.Version == 0 {
			missingFields = append(missingFields, "dataset.version")
		}

		if filter.Dataset.Edition == "" {
			missingFields = append(missingFields, "dataset.edition")
		}

		if filter.Dataset.ID == "" {
			missingFields = append(missingFields, "dataset.id")
		}
	} else {
		missingFields = append(missingFields, "dataset.version", "dataset.edition", "dataset.id")
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}

// DatasetDimensionResults represents a structure for a list of dimensions
type DatasetDimensionResults struct {
	Items []DatasetDimension `json:"items"`
}

// DatasetDimension represents an overview for a single dimension. This includes a link to the code list API
// which provides metadata about the dimension and all possible values.
type DatasetDimension struct {
	Name string `bson:"name,omitempty"          json:"dimension,omitempty"`
}

// ValidateFilterDimensions checks the selected filter dimension
// are valid for a version of a dataset
func ValidateFilterDimensions(filterDimensions []Dimension, datasetDimensions *DatasetDimensionResults) error {
	dimensionNames := make(map[string]int)
	for _, datasetDimension := range datasetDimensions.Items {
		dimensionNames[datasetDimension.Name] = 1
	}

	var incorrectDimensions []string
	for _, filterDimension := range filterDimensions {
		if _, ok := dimensionNames[filterDimension.Name]; !ok {
			incorrectDimensions = append(incorrectDimensions, filterDimension.Name)
		}
	}

	if incorrectDimensions != nil {
		return fmt.Errorf("Bad request - incorrect dimensions chosen: %v", incorrectDimensions)
	}

	return nil
}

// DatasetDimensionOptionResults represents a structure for a list of dimension options
type DatasetDimensionOptionResults struct {
	Items []PublicDimensionOption `json:"items"`
}

// PublicDimensionOption hides values which are only used by interval services
type PublicDimensionOption struct {
	Name   string `bson:"name,omitempty"           json:"dimension"`
	Label  string `bson:"label,omitempty"          json:"label"`
	Option string `bson:"option,omitempty"         json:"option"`
}

// ValidateFilterDimensionOptions checks the selected filter dimension options
// are valid for a dimension of a single version of a dataset
func ValidateFilterDimensionOptions(filterDimensionOptions []string, datasetDimensionOptions *DatasetDimensionOptionResults) []string {
	dimensionOptions := make(map[string]int)
	for _, datasetOption := range datasetDimensionOptions.Items {
		dimensionOptions[datasetOption.Option] = 1
	}

	var incorrectDimensionOptions []string
	for _, filterOption := range filterDimensionOptions {
		if _, ok := dimensionOptions[filterOption]; !ok {
			incorrectDimensionOptions = append(incorrectDimensionOptions, filterOption)
		}
	}

	return incorrectDimensionOptions
}

// ValidateFilterOutputUpdate checks the content of the filter structure
func (filter *Filter) ValidateFilterOutputUpdate(currentFilter *Filter) error {

	// Only downloads, events and state can be updated, any attempt to update other
	// fields will result in an error of forbidden

	var forbiddenFields []string

	if filter.Dataset != nil {
		if filter.Dataset.ID != "" {
			forbiddenFields = append(forbiddenFields, "dataset.id")
		}

		if filter.Dataset.Edition != "" {
			forbiddenFields = append(forbiddenFields, "dataset.edition")
		}

		if filter.Dataset.Version != 0 {
			forbiddenFields = append(forbiddenFields, "dataset.version")
		}
	}

	if filter.InstanceID != "" {
		forbiddenFields = append(forbiddenFields, "instance_id")
	}

	if filter.Dimensions != nil {
		forbiddenFields = append(forbiddenFields, "dimensions")
	}

	if filter.FilterID != "" {
		forbiddenFields = append(forbiddenFields, "filter_id")
	}

	if currentFilter.Published != nil && currentFilter.Published == &Published && currentFilter.Downloads != nil {
		if filter.Downloads != nil {
			if filter.Downloads.CSV != nil {
				// If version for filter output is published and filter output has a
				// public link, do not allow any updates to download item (csv)
				var hasCSVPublicDownload bool
				if currentFilter.Downloads.CSV != nil && currentFilter.Downloads.CSV.Public != "" {
					hasCSVPublicDownload = true
					forbiddenFields = append(forbiddenFields, "downloads.csv")
				}

				// If version for filter output is published, do not allow updates to create
				// csv private link unless there are no downloads
				if !hasCSVPublicDownload && filter.Downloads.CSV.Private != "" {
					forbiddenFields = append(forbiddenFields, "downloads.csv.private")
				}
			}

			if filter.Downloads.XLS != nil {
				// If version for filter output is published and filter output has a
				// public link, do not allow any updates to download item (xls)
				var hasXLSPublicDownload bool
				if currentFilter.Downloads.XLS != nil && currentFilter.Downloads.XLS.Public != "" {
					hasXLSPublicDownload = true
					forbiddenFields = append(forbiddenFields, "downloads.xls")
				}

				// If version for filter output is published, do not allow updates to create
				// xls private link unless there are no downloads
				if !hasXLSPublicDownload && filter.Downloads.XLS.Private != "" {
					forbiddenFields = append(forbiddenFields, "downloads.xls.private")
				}
			}
		}
	}

	if forbiddenFields != nil {
		return fmt.Errorf("Forbidden from updating the following fields: %v", forbiddenFields)
	}

	return nil
}

// ValidateFilterBlueprintUpdate checks the content of the filter structure for
// changes against the dataset
func ValidateFilterBlueprintUpdate(filter *Filter) error {

	// Only events and dataset version can be updated, any attempt to update the
	// dataset id or edition will result in an error of bad request

	var forbiddenFields []string

	if filter.Dataset != nil {
		if filter.Dataset.ID != "" {
			forbiddenFields = append(forbiddenFields, "dataset.id")
		}

		if filter.Dataset.Edition != "" {
			forbiddenFields = append(forbiddenFields, "dataset.edition")
		}

		if forbiddenFields != nil {
			return fmt.Errorf("Forbidden from updating the following fields: %v", forbiddenFields)
		}
	}

	return nil
}

// CreateFilter manages the creation of a filter from a reader
func CreateFilter(reader io.Reader) (*Filter, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ErrorReadingBody
	}

	var filter Filter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, ErrorParsingBody
	}

	// This should be the last check before returning filter, as we want to check if a empty json `{}` message has been
	// sent from the client
	if len(bytes) == 2 {
		return &filter, ErrorNoData
	}

	return &filter, nil
}

// CreateDimensionOptions manages the creation of options for a dimension from a reader
func CreateDimensionOptions(reader io.Reader) ([]string, error) {
	var dimension Dimension

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ErrorReadingBody
	}

	if string(bytes) == "" {
		return dimension.Options, nil
	}

	err = json.Unmarshal(bytes, &dimension)
	if err != nil {
		return nil, ErrorParsingBody
	}

	return dimension.Options, nil
}

// CreateNewFilter manages the creation of a filter blueprint being updated (new filter)
func CreateNewFilter(reader io.Reader) (*NewFilter, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ErrorReadingBody
	}

	var filter NewFilter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, ErrorParsingBody
	}

	// This should be the last check before returning filter
	if len(bytes) == 2 {
		return &filter, ErrorNoData
	}

	return &filter, nil
}
