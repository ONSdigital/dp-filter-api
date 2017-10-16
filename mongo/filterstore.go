package mongo

import (
	"errors"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const Database = "filters"
const FiltersCollection = "filters"

const Submitted = "submitted"

var NotFound = errors.New("not found")
var NotAuthorised = errors.New("Not authorised")

// FilterStore containing all filter jobs stored in mongodb
type FilterStore struct {
	Session *mgo.Session
}

// CreateFilterStore which can store, update and fetch filter jobs
func CreateFilterStore(url string) (*FilterStore, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return &FilterStore{Session: session}, nil
}

// AddFilter to the data store
func (s *FilterStore) AddFilter(host string, filter *models.Filter) (*models.Filter, error) {
	session := s.Session.Copy()
	filter.FilterID = uuid.NewV4().String()
	err := session.DB(Database).C(FiltersCollection).Insert(filter)
	if err != nil {
		return nil, err
	}
	return filter, nil
}

// GetFilter returns a single filter, if not found return an error
func (s *FilterStore) GetFilter(filterID string) (*models.Filter, error) {
	session := s.Session.Copy()
	query := bson.M{"filter_job_id": filterID}
	var result models.Filter
	err := session.DB(Database).C(FiltersCollection).Find(query).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, NotFound
		}
		return nil, err
	}
	validateFilter(&result)
	return &result, nil
}

// UpdateFilter replaces the stored filter properties
func (s *FilterStore) UpdateFilter(isAuthenticated bool, UpdatedFilter *models.Filter) error {
	session := s.Session.Copy()

	currentFilter, err := s.GetFilter(UpdatedFilter.FilterID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}

	if currentFilter.State == Submitted && !isAuthenticated {
		return NotAuthorised
	}

	query := bson.M{"filter_job_id": UpdatedFilter.FilterID}
	err = session.DB(Database).C(FiltersCollection).Update(query, &UpdatedFilter)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}
	return nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(*models.AddDimension) error {
	return nil
}

// AddFilterDimensionOption to a filter
func (s *FilterStore) AddFilterDimensionOption(*models.AddDimensionOption) error {
	return nil
}

// GetFilterDimensions returns a list of dimensions from a filter, if the filter is not found an error is returned
func (s *FilterStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	return nil, nil
}

// GetFilterDimension return a single dimension
func (s *FilterStore) GetFilterDimension(filterID string, name string) error {
	return nil
}

// GetFilterDimensionOptions return a list of dimension options
func (s *FilterStore) GetFilterDimensionOptions(filterID string, name string) ([]models.DimensionOption, error) {
	return nil, nil
}

// GetFilterDimensionOption return a single dimension option
func (s *FilterStore) GetFilterDimensionOption(filterID string, name string, option string) error {
	return nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID string, name string) error {
	return nil
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string) error {
	return nil
}

func validateFilter(filter *models.Filter) {
	// Make sure all empty arrays are initialized
	if filter.Dimensions == nil {
		filter.Dimensions = []models.Dimension{}
	}
	if filter.Events.Info == nil {
		filter.Events.Info = []models.EventItem{}
	}
	if filter.Events.Error == nil {
		filter.Events.Error = []models.EventItem{}
	}
}
