package mongo

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"gopkg.in/mgo.v2"
)

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
func (s *FilterStore) AddFilter(host string, filter *models.Filter) (models.Filter, error) {
	return models.Filter{}, nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(*models.AddDimension) error {
	return nil
}

// AddFilterDimensionOption to a filter
func (s *FilterStore) AddFilterDimensionOption(*models.AddDimensionOption) error {
	return nil
}

// GetFilter returns a single filter, if not found return an error
func (s *FilterStore) GetFilter(filterID string) (models.Filter, error) {
	return models.Filter{}, nil
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

// UpdateFilter replaces the stored filter properties
func (s *FilterStore) UpdateFilter(isAuthenticated bool, filter *models.Filter) error {
	return nil
}
