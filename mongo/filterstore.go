package mongo

import (
	"errors"
	"fmt"
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
var DimensionNotFound = errors.New("Dimension not found")
var Forbidden = errors.New("Forbidden")

// FilterStore containing all filter jobs stored in mongodb
type FilterStore struct {
	Session *mgo.Session
	host    string
}

// CreateFilterStore which can store, update and fetch filter jobs
func CreateFilterStore(url, host string) (*FilterStore, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return &FilterStore{Session: session, host: host}, nil
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

// GetFilterDimensions returns a list of dimensions from a filter, if the filter is not found an error is returned
func (s *FilterStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	session := s.Session.Copy()
	query := bson.M{"filter_job_id": filterID}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter
	err := session.DB(Database).C(FiltersCollection).Find(query).Select(dimensionSelect).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, NotFound
		}
		return nil, err
	}
	return result.Dimensions, nil
}

// GetFilterDimension return a single dimension
func (s *FilterStore) GetFilterDimension(filterID string, name string) error {
	session := s.Session.Copy()
	queryFilter := bson.M{"filter_job_id": filterID}
	queryDimension := bson.M{"filter_job_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter
	err := session.DB(Database).C(FiltersCollection).Find(queryFilter).Select(dimensionSelect).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}
	err = session.DB(Database).C(FiltersCollection).Find(queryDimension).Select(dimensionSelect).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return DimensionNotFound
		}
		return err
	}
	return nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(dimension *models.AddDimension) error {
	err := s.checkFilterState(dimension.FilterID)
	if err != nil {
		return err
	}

	session := s.Session.Copy()
	queryFilter := bson.M{"filter_job_id": dimension.FilterID}
	url := fmt.Sprintf("%s/filter/%s/dimensions/%s", s.host, dimension.FilterID, dimension.Name)
	d := models.Dimension{Name: dimension.Name, Options: dimension.Options, DimensionURL: url}
	update := bson.M{"$push": bson.M{"dimensions": d}}
	err = session.DB(Database).C(FiltersCollection).Update(queryFilter, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}
	return nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID string, name string) error {
	err := s.checkFilterState(filterID)
	if err != nil {
		return err
	}
	session := s.Session.Copy()
	queryFilter := bson.M{"filter_job_id": filterID}
	update := bson.M{"$pull": bson.M{"dimensions": bson.M{"name": name}}}
	err = session.DB(Database).C(FiltersCollection).Update(queryFilter, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}
	return nil
}

// AddFilterDimensionOption to a filter
func (s *FilterStore) AddFilterDimensionOption(*models.AddDimensionOption) error {
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

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string) error {
	return nil
}

func (s *FilterStore) checkFilterState(filterID string) error {
	filter, err := s.GetFilter(filterID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return NotFound
		}
		return err
	}

	if filter.State == Submitted {
		return Forbidden
	}
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
