package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// Filter represents a structure for a filter job
type Filter struct {
	DataSetFilterID  string      `json:"dataset_filter_id"`
	DimensionListURL string      `json:"dimension_list_url,omitempty"`
	Dimensions       []Dimension `json:"dimensions,omitempty"`
	Downloads        Downloads   `json:"downloads,omitempty"`
	Events           Events      `json:"events,omitempty"`
	FilterID         string      `json:"id,omitempty"`
	State            string      `json:"state,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	DimensionValueURL string   `json:"dimension_value_url,omitempty"`
	Name              string   `json:"name,omitempty"`
	Values            []string `json:"values,omitempty"`
}

// Downloads represents a list of file types possible to download
type Downloads struct {
	CSV  DownloadItem `json:"csv,omitempty"`
	JSON DownloadItem `json:"json,omitempty"`
	XLS  DownloadItem `json:"xls,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Events represents a list of array objects containing event information against the filter job
type Events struct {
	Error []EventItem `json:"error,omitempty"`
	Info  []EventItem `json:"info,omitempty"`
}

// EventItem represents an event object containing event information
type EventItem struct {
	Message string `json:"message,omitempty"`
	Time    string `json:"time,omitempty"`
	Type    string `json:"type,omitempty"`
}

// Validate checks the content of the filter structure
func (filter *Filter) Validate() error {
	if filter.State == "" {
		filter.State = "created"
	}

	var missingFields []string

	if filter.DataSetFilterID == "" {
		missingFields = append(missingFields, "dataset_filter_id")
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
		return nil, fmt.Errorf("Failed to read message body")
	}
	var filter Filter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json body")
	}

	return &filter, nil
}
