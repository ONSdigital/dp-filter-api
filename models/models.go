package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// Filter represents a structure for a filter job
type Filter struct {
	InstanceID       string      `bson:"instance_id"          json:"instance_id"`
	DimensionListURL string      `bson:"dimension_list_url"   json:"dimension_list_url,omitempty"`
	Dimensions       []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Downloads        Downloads   `bson:"downloads,omitempty"  json:"downloads,omitempty"`
	Events           Events      `bson:"events,omitempty"     json:"events,omitempty"`
	FilterID         string      `bson:"filter_job_id"        json:"filter_job_id,omitempty"`
	State            string      `bson:"state"                json:"state,omitempty"`
	Links            LinkMap     `bson:"links"                json:"links,omitempty"`
}

// LinkMap contains a named LinkObject for each link to other resources
type LinkMap struct {
	Version LinkObject `bson:"version" json:"version,omitempty"`
}

// LinkObject represents a generic structure for all links
type LinkObject struct {
	ID   string `bson:"id"   json:"id,omitempty"`
	HRef string `bson:"href" json:"href,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	DimensionURL string   `bson:"dimension_url"           json:"dimension_url"`
	Name         string   `bson:"name"                    json:"name"`
	Options      []string `bson:"options"                 json:"options"`
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

// Validate checks the content of the filter structure
func (filter *Filter) Validate() error {
	if filter.State == "" {
		filter.State = "created"
	}

	var missingFields []string

	if filter.InstanceID == "" {
		missingFields = append(missingFields, "instance_id")
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}

// CreateFilter manages the creation of a filter from a reader
func CreateFilter(reader io.Reader) (*Filter, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var filter Filter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}

	return &filter, nil
}

// CreateDimensionOptions manages the creation of options for a dimension from a reader
func CreateDimensionOptions(reader io.Reader) ([]string, error) {
	var dimension Dimension

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}

	if string(bytes) == "" {
		return dimension.Options, nil
	}

	err = json.Unmarshal(bytes, &dimension)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}

	return dimension.Options, nil
}
