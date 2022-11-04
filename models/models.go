package models

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	dprequest "github.com/ONSdigital/dp-net/request"
	"go.mongodb.org/mongo-driver/bson"
)

// A list of states
const (
	CreatedState   = "created"
	CompletedState = "completed"
)

var (
	Unpublished = false
	Published   = true
)

// Dataset contains the uniique identifiers that make a dataset unique
type Dataset struct {
	ID              string `bson:"id"        json:"id"`
	Edition         string `bson:"edition"   json:"edition"`
	Version         int    `bson:"version"   json:"version"`
	LowestGeography string `bson:"lowest_geography" json:"lowest_geography"`
}

// NewFilter creates a filter by using the dataset id, edition and version
type NewFilter struct {
	Dataset    *Dataset    `bson:"dataset" json:"dataset"`
	Dimensions []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
}

// Filter represents a structure for a filter job
type Filter struct {
	UniqueTimestamp primitive.Timestamp `bson:"unique_timestamp,omitempty" json:"-"`
	LastUpdated     time.Time           `bson:"last_updated"               json:"-"`
	ETag            string              `bson:"e_tag"                      json:"-"`

	ID         string      `bson:"id"                    json:"id,omitempty"`
	Dataset    *Dataset    `bson:"dataset"              json:"dataset"`
	InstanceID string      `bson:"instance_id"          json:"instance_id"`
	Dimensions []Dimension `bson:"dimensions,omitempty" json:"dimensions,omitempty"`
	Downloads  *Downloads  `bson:"downloads,omitempty"  json:"downloads,omitempty"`
	Events     []*Event    `bson:"events,omitempty"     json:"events,omitempty"`
	FilterID   string      `bson:"filter_id"            json:"filter_id,omitempty"`
	State      string      `bson:"state,omitempty"      json:"state,omitempty"`
	Published  *bool       `bson:"published,omitempty"  json:"published,omitempty"`
	Links      LinkMap     `bson:"links"                json:"links,omitempty"`
	Type       string      `bson:"type,omitempty"       json:"type,omitempty"`
}

// Hash generates a SHA-1 hash of the filter struct. SHA-1 is not cryptographically safe,
// but it has been selected for performance as we are only interested in uniqueness.
// ETag field value is ignored when generating a hash.
// An optional byte array can be provided to append to the hash.
// This can be used, for example, to calculate a hash of this filter and an update applied to it.
func (f *Filter) Hash(extraBytes []byte) (string, error) {
	h := sha1.New()

	// copy by value to ignore ETag without affecting f
	f2 := *f
	f2.ETag = ""

	filterBytes, err := bson.Marshal(f2)
	if err != nil {
		return "", err
	}

	if _, err := h.Write(append(filterBytes, extraBytes...)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// LinkMap contains a named *LinkObject for each link to other resources
type LinkMap struct {
	Dimensions      *LinkObject `bson:"dimensions"                 json:"dimensions,omitempty"`
	FilterOutput    *LinkObject `json:"filter_output,omitempty"`
	FilterBlueprint *LinkObject `bson:"filter_blueprint,omitempty" json:"filter_blueprint,omitempty"`
	Self            *LinkObject `bson:"self"                       json:"self,omitempty"`
	Version         *LinkObject `bson:"version"                    json:"version,omitempty"`
}

// LinkObject represents a generic structure for all links
type LinkObject struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty"`
	HRef string `bson:"href"         json:"href,omitempty"`
}

// Dimension represents an object containing a list of dimension values and the dimension name
type Dimension struct {
	URL        string   `bson:"dimension_url,omitempty" json:"dimension_url,omitempty"`
	Name       string   `bson:"name"                    json:"name"`
	Options    []string `bson:"options,omitempty"       json:"options"`
	IsAreaType *bool    `bson:"is_area_type,omitempty"  json:"is_area_type,omitempty"`
}

type UpdateDimensionResponse struct {
	ID    string             `json:"id"`
	Name  string             `json:"name"`
	Label string             `json:"label"`
	Links DimensionItemLinks `json:"links"`
}

type DimensionItemLinks struct {
	Filter  LinkObject `json:"filter"`
	Options LinkObject `json:"options"`
	Self    LinkObject `json:"self"`
}

// EncodedOptions returns the list of options for this dimension after escaping the values for URL query paramters
func (d *Dimension) EncodedOptions() []string {
	encodedIDs := make([]string, len(d.Options))
	for i, op := range d.Options {
		encodedIDs[i] = url.QueryEscape(op)
	}
	return encodedIDs
}

// PublicDimension represents information about a single dimension as served by /dimensions and /dimensions/<id>
type PublicDimension struct {
	Name  string                  `bson:"name"                    json:"name"`
	Links *PublicDimensionLinkMap `bson:"links"                   json:"links"`
}

type PublicDimensions struct {
	Items      []*PublicDimension `json:"items"`
	Count      int                `json:"count"`
	Offset     int                `json:"offset"`
	Limit      int                `json:"limit"`
	TotalCount int                `json:"total_count"`
}

// PublicDimensionLinkMap is the links map for the PublicDimension structure
type PublicDimensionLinkMap struct {
	Self    *LinkObject `bson:"self"                 json:"self"`
	Filter  *LinkObject `bson:"filter"               json:"filter"`
	Options *LinkObject `bson:"options"              json:"options,omitempty"`
}

// PublicDimensionOption represents information about a single dimension option as served by /options and /options/<id>
type PublicDimensionOption struct {
	Links  *PublicDimensionOptionLinkMap `bson:"links"               json:"links"`
	Option string                        `bson:"option"              json:"option"`
}

// PublicDimensionOptions represents information about a set of dimension options
type PublicDimensionOptions struct {
	Items      []*PublicDimensionOption `json:"items"`
	Count      int                      `json:"count"`
	Offset     int                      `json:"offset"`
	Limit      int                      `json:"limit"`
	TotalCount int                      `json:"total_count"`
}

// PublicDimensionOptionLinkMap is the links map for the PublicDimensionOption structure
type PublicDimensionOptionLinkMap struct {
	Self      *LinkObject `bson:"self"                 json:"self"`
	Filter    *LinkObject `bson:"filter"               json:"filter"`
	Dimension *LinkObject `bson:"dimension"            json:"dimension"`
}

// Downloads represents a list of file types possible to download
type Downloads struct {
	CSV *DownloadItem `bson:"csv,omitempty"  json:"csv,omitempty"`
	XLS *DownloadItem `bson:"xls,omitempty"  json:"xls,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Skipped bool   `bson:"skipped,omitempty" json:"skipped,omitempty"`
	HRef    string `bson:"href,omitempty"    json:"href,omitempty"`
	Private string `bson:"private,omitempty" json:"private,omitempty"`
	Public  string `bson:"public,omitempty"  json:"public,omitempty"`
	Size    string `bson:"size,omitempty"    json:"size,omitempty"`
}

// A list of errors returned from package
var (
	ErrorReadingBody = errors.New("failed to read message body")
	ErrorParsingBody = errors.New("failed to parse json body")
	ErrorNoData      = errors.New("bad request - missing data in body")
)

// DuplicateDimensionError is returned if a request contains a duplicate dimension
type DuplicateDimensionError struct {
	duplicateDimension string
}

func (e DuplicateDimensionError) Error() string {
	return fmt.Sprintf("Bad request - duplicate dimension found: %s", e.duplicateDimension)
}

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

		if filter.Dataset.LowestGeography == "" {
			missingFields = append(missingFields, "dataset.lowest_geography")
		}
	} else {
		missingFields = append(missingFields, "dataset.version", "dataset.edition", "dataset.id", "dataset.lowest_geography")
	}

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	return nil
}

// ValidateFilterDimensions checks the selected filter dimension
// are valid for a version of a dataset
func ValidateFilterDimensions(filterDimensions []Dimension, dimensions *dataset.VersionDimensions) error {
	dimensionNames := make(map[string]int)
	for _, dimension := range dimensions.Items {
		dimensionNames[dimension.Name] = 1
	}

	var incorrectDimensions []string
	for _, filterDimension := range filterDimensions {
		if _, ok := dimensionNames[filterDimension.Name]; !ok {
			incorrectDimensions = append(incorrectDimensions, filterDimension.Name)
		}
	}

	if incorrectDimensions != nil {
		return fmt.Errorf("incorrect dimensions chosen: %v", incorrectDimensions)
	}

	return nil
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

	if currentFilter.Published != nil && *currentFilter.Published == Published && currentFilter.Downloads != nil {
		if filter.Downloads != nil {
			if filter.Downloads.CSV != nil && currentFilter.Downloads.CSV != nil {
				// If version for filter output is published and filter output has a
				// public link, do not allow any updates to download item (csv)
				var hasCSVPublicDownload bool
				if currentFilter.Downloads.CSV.Public != "" {
					hasCSVPublicDownload = true
					forbiddenFields = append(forbiddenFields, "downloads.csv")
				}

				// If version for filter output is published, do not allow updates to create
				// csv private link unless there are no downloads
				if !hasCSVPublicDownload && filter.Downloads.CSV.Private != "" {
					forbiddenFields = append(forbiddenFields, "downloads.csv.private")
				}
			}

			if filter.Downloads.XLS != nil && currentFilter.Downloads.XLS != nil {

				// If version for filter output is published and filter output has a
				// public link, do not allow any updates to download item (xls)
				var hasXLSPublicDownload bool
				if currentFilter.Downloads.XLS.Public != "" {
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
		return fmt.Errorf("forbidden from updating the following fields: %v", forbiddenFields)
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
			return fmt.Errorf("forbidden from updating the following fields: %v", forbiddenFields)
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

// CreatePatches manages the creation of an array of patch structs from the provided reader, and validates them
func CreatePatches(reader io.Reader) ([]dprequest.Patch, error) {
	patches := []dprequest.Patch{}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return []dprequest.Patch{}, ErrorReadingBody
	}

	err = json.Unmarshal(bytes, &patches)
	if err != nil {
		return []dprequest.Patch{}, ErrorParsingBody
	}

	for _, patch := range patches {
		if err := patch.Validate(dprequest.OpAdd, dprequest.OpRemove); err != nil {
			return []dprequest.Patch{}, err
		}
	}
	return patches, nil
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

	dimensionValidator := make(map[string]bool)
	for _, dimension := range filter.Dimensions {
		if dimensionValidator[dimension.Name] {
			return nil, DuplicateDimensionError{dimension.Name}
		}
		dimensionValidator[dimension.Name] = true
	}

	// This should be the last check before returning filter
	if len(bytes) == 2 {
		return &filter, ErrorNoData
	}

	return &filter, nil
}
