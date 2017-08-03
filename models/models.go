package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// Filter represents a structure for a filter job
type Filter struct {
	DataSet    string      `json:"dataset"`
	Edition    string      `json:"edition"`
	FilterID   string      `json:"id,omitempty"`
	State      string      `json:"state,omitempty"`
	Version    string      `json:"version"`
	Dimensions []Dimension `json:"dimensions,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}

// ValidateFilterCreation checks the content of the filter structure
func (filter *Filter) ValidateFilterCreation() error {
	if filter.State == "" {
		filter.State = "created"
	}

	var missingFields []string

	if filter.DataSet == "" {
		missingFields = append(missingFields, "dataset")
	}

	if filter.Edition == "" {
		missingFields = append(missingFields, "edition")
	}

	if filter.Version == "" {
		missingFields = append(missingFields, "version")
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
