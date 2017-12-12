package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

// Filter represents a structure for a filter job
type Filter struct {
	InstanceID  string      `bson:"instance_id"          json:"instance_id"`
	Dimensions  []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Downloads   *Downloads  `bson:"downloads,omitempty"  json:"downloads,omitempty"`
	Events      Events      `bson:"events,omitempty"     json:"events,omitempty"`
	FilterID    string      `bson:"filter_id"            json:"filter_id,omitempty"`
	State       string      `bson:"state,omitempty"      json:"state,omitempty"`
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
	CSV  DownloadItem `bson:"csv"  json:"csv"`
	JSON DownloadItem `bson:"json" json:"json"`
	XLS  DownloadItem `bson:"xls"  json:"xls"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `bson:"size,omitempty" json:"size"`
	URL  string `bson:"url,omitempty"  json:"url"`
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

// AddDimension represents dimension information for storing a list of options for a dimension
type AddDimension struct {
	FilterID string
	Name     string
	Options  []string
}

// AddDimensionOption represents dimension option information for storing
// an individual option for a given filter job dimension
type AddDimensionOption struct {
	FilterID string
	Name     string
	Option   string
}

// DimensionOption represents dimension option information
type DimensionOption struct {
	DimensionOptionURL string `json:"dimension_option_url"`
	Option             string `json:"option"`
}

var (
	ErrorReadingBody = errors.New("Failed to read message body")
	ErrorParsingBody = errors.New("Failed to parse json body")
	ErrorNoData      = errors.New("Bad request - Missing data in body")
)

// ValidateFilterBlueprint checks the content of the filter structure
func (filter *Filter) ValidateFilterBlueprint() error {
	// FilterBluePrint should not have state
	filter.State = ""

	var missingFields []string

	if filter.InstanceID == "" {
		missingFields = append(missingFields, "instance_id")
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
	dimensionNames := make(map[string]string)
	for _, datasetDimension := range datasetDimensions.Items {
		dimensionNames[datasetDimension.Name] = datasetDimension.Name
	}

	var incorrectDimensions []string
	for _, filterDimension := range filterDimensions {
		if dimensionNames[filterDimension.Name] != filterDimension.Name {
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
	dimensionOptions := make(map[string]string)
	for _, datasetOption := range datasetDimensionOptions.Items {
		dimensionOptions[datasetOption.Option] = datasetOption.Option
	}

	var incorrectDimensionOptions []string
	for _, filterOption := range filterDimensionOptions {
		if dimensionOptions[filterOption] != filterOption {
			incorrectDimensionOptions = append(incorrectDimensionOptions, filterOption)
		}
	}

	return incorrectDimensionOptions
}

// ValidateFilterOutputIpdate checks the content of the filter structure
func (filter *Filter) ValidateFilterOutputUpdate() error {

	// Only downloads, events and state can be updated, any attempt to update other
	// fields will result in an error of forbidden

	var forbiddenFields []string

	if filter.InstanceID != "" {
		forbiddenFields = append(forbiddenFields, "instance_id")
	}

	if filter.Dimensions != nil {
		forbiddenFields = append(forbiddenFields, "dimensions")
	}

	if filter.FilterID != "" {
		forbiddenFields = append(forbiddenFields, "filter_id")
	}

	if forbiddenFields != nil {
		return fmt.Errorf("Forbidden from updating the following fields: %v", forbiddenFields)
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

	// This should be the last check before returning filter
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
