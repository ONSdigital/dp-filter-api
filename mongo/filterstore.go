package mongo

import (
	"gopkg.in/mgo.v2"
	"github.com/ONSdigital/dp-filter-api/models"
)

type FilterStore struct {
	Session *mgo.Session
}

func CreateFilterStore(url string) (*FilterStore, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return &FilterStore{Session: session}, nil
}

func (s *FilterStore) AddFilter(host string, filter *models.Filter) (models.Filter, error) {
	return models.Filter{}, nil
}

func (s *FilterStore) AddFilterDimension(*models.AddDimension) error {
	return nil
}

func (s *FilterStore) AddFilterDimensionOption(*models.AddDimensionOption) error {
	return nil
}

func (s *FilterStore) GetFilter(filterID string) (models.Filter, error) {
	return models.Filter{}, nil
}
func (s *FilterStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	return nil, nil
}

func (s *FilterStore) GetFilterDimension(filterID string, name string) error {
	return nil
}

func (s *FilterStore) GetFilterDimensionOptions(filterID string, name string) ([]models.DimensionOption, error) {
	return nil, nil
}

func (s *FilterStore) GetFilterDimensionOption(filterID string, name string, option string) error {
	return nil
}

func (s *FilterStore) RemoveFilterDimension(filterID string, name string) error {
	return nil
}

func (s *FilterStore) RemoveFilterDimensionOption(filterId string, name string, option string) error {
	return nil
}

func (s *FilterStore) UpdateFilter(isAuthenticated bool, filter *models.Filter) error {
	return nil
}
