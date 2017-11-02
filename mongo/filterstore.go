package mongo

import (
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-filter-api/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Database variables
const (
	database                = "filters"
	filtersCollection       = "filters"
	filterOutputsCollection = "filterOutputs"
)

// Filter job states
const (
	submitted = "submitted"
	completed = "completed"
)

// Error codes
var (
	errNotAuthorised = errors.New("Not authorised")
	errForbidden     = errors.New("Forbidden")

	errNotFound                  = errors.New("Not found")
	errDimensionNotFound         = errors.New("Dimension not found")
	errOptionNotFound            = errors.New("Option not found")
	errFilterOrDimensionNotFound = errors.New("Bad request - filter or dimension not found")
	errFilterOutputNotFound      = errors.New("Filter output not found")

	errFilterBadRequest    = errors.New("Bad request - filter blueprint not found")
	errDimensionBadRequest = errors.New("Bad request - filter dimension not found")
)

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
	defer session.Close()

	if err := session.DB(database).C(filtersCollection).Insert(filter); err != nil {
		return nil, err
	}

	validateFilter(filter)
	return filter, nil
}

// GetFilter returns a single filter, if not found return an error
func (s *FilterStore) GetFilter(filterID string) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	query := bson.M{"filter_id": filterID}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(query).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errNotFound
		}
		return nil, err
	}

	validateFilter(&result)
	return &result, nil
}

// UpdateFilter replaces the stored filter properties
func (s *FilterStore) UpdateFilter(updatedFilter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	update := createUpdateFilterBlueprint(updatedFilter, time.Now())

	selector := bson.M{"filter_id": updatedFilter.FilterID}
	if err := session.DB(database).C(filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	return nil
}

// GetFilterDimensions returns a list of dimensions from a filter, if the filter is not found an error is returned
func (s *FilterStore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	session := s.Session.Copy()
	defer session.Close()

	query := bson.M{"filter_id": filterID}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(query).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errNotFound
		}
		return nil, err
	}

	return result.Dimensions, nil
}

// GetFilterDimension return a single dimension
func (s *FilterStore) GetFilterDimension(filterID string, name string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_id": filterID}
	queryDimension := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(queryFilter).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return errFilterBadRequest
		}
		return err
	}

	if err := session.DB(database).C(filtersCollection).Find(queryDimension).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return errDimensionNotFound
		}
		return err
	}

	return nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(dimension *models.AddDimension) error {
	if err := s.checkFilterState(dimension.FilterID); err != nil {
		return err
	}

	session := s.Session.Copy()
	defer session.Close()

	list, err := s.GetFilterDimensions(dimension.FilterID)
	if err != nil && err != errNotFound {
		return errFilterBadRequest
	}

	url := fmt.Sprintf("%s/filters/%s/dimensions/%s", s.host, dimension.FilterID, dimension.Name)
	d := models.Dimension{Name: dimension.Name, Options: dimension.Options, URL: url}

	var found bool
	for i, item := range list {
		if item.Name == d.Name {
			found = true
			list[i] = d
			break
		}
	}

	if !found {
		list = append(list, d)
	}

	queryFilter := bson.M{"filter_id": dimension.FilterID}
	update := bson.M{"$set": bson.M{"dimensions": list}}

	if err := session.DB(database).C(filtersCollection).Update(queryFilter, update); err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	return nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID string, name string) error {
	if err := s.checkFilterState(filterID); err != nil {
		if err == errNotFound {
			return errFilterBadRequest
		}
		return err
	}

	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_id": filterID}
	update := bson.M{"$pull": bson.M{"dimensions": bson.M{"name": name}}}

	info, err := session.DB(database).C(filtersCollection).UpdateAll(queryFilter, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	if info.Updated == 0 {
		return errDimensionNotFound
	}

	return nil
}

// AddFilterDimensionOption to a filter
func (s *FilterStore) AddFilterDimensionOption(newOption *models.AddDimensionOption) error {
	if err := s.checkFilterState(newOption.FilterID); err != nil {
		if err == errNotFound {
			return errFilterBadRequest
		}
		return err
	}

	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": newOption.FilterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": newOption.Name}}}
	update := bson.M{"$addToSet": bson.M{"dimensions.$.options": newOption.Option}}

	if err := session.DB(database).C(filtersCollection).Update(queryOptions, update); err != nil {
		if err == mgo.ErrNotFound {
			return errDimensionNotFound
		}
		return err
	}

	return nil
}

// GetFilterDimensionOptions return a list of dimension options
func (s *FilterStore) GetFilterDimensionOptions(filterID string, name string) ([]models.DimensionOption, error) {
	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_id": filterID}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(queryFilter).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errFilterBadRequest
		}
		return nil, err
	}

	for _, dimension := range result.Dimensions {
		var options []models.DimensionOption

		if dimension.Name == name {
			for _, option := range dimension.Options {
				url := fmt.Sprintf("%s/filter/%s/dimensions/%s/option/%s", s.host, filterID, dimension.Name, option)
				dimensionOption := models.DimensionOption{Option: option, DimensionOptionURL: url}
				options = append(options, dimensionOption)
			}

			return options, nil
		}
	}

	return nil, errDimensionNotFound
}

// GetFilterDimensionOption return a single dimension option
func (s *FilterStore) GetFilterDimensionOption(filterID string, name string, option string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryDimension := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(queryDimension).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return errFilterOrDimensionNotFound
		}
		return err
	}

	for _, d := range result.Dimensions {
		if d.Name == name {
			for _, o := range d.Options {
				if o == option {
					return nil
				}
			}
		}
	}

	return errOptionNotFound
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string) error {
	// Check if filter exists
	dimensions, err := s.GetFilterDimensions(filterID)
	if err != nil {
		if err == errNotFound {
			return errFilterBadRequest
		}
		return err
	}

	var hasDimension bool

	// Check if dimension exists
	for _, dimension := range dimensions {
		if dimension.Name == name {
			hasDimension = true
			break
		}
	}

	if !hasDimension {
		return errDimensionBadRequest
	}

	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update := bson.M{"$pull": bson.M{"dimensions.$.options": option}}

	info, err := session.DB(database).C(filtersCollection).UpdateAll(queryOptions, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errOptionNotFound
		}
		return err
	}

	// document was match but nothing was removed
	if info.Updated == 0 {
		return errOptionNotFound
	}

	return nil
}

func (s *FilterStore) CreateFilterOutput(filter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	if err := session.DB(database).C(filterOutputsCollection).Insert(filter); err != nil {
		return err
	}

	return nil
}

func (s *FilterStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	query := bson.M{"filter_id": filterID}
	var result *models.Filter

	if err := session.DB(database).C(filterOutputsCollection).Find(query).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errFilterOutputNotFound
		}

		return nil, err
	}

	return result, nil
}

func (s *FilterStore) UpdateFilterOutput(filter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	update := createUpdateFilterOutput(filter, time.Now())

	if err := session.DB(database).C(filterOutputsCollection).
		Update(bson.M{"filter_id": filter.FilterID}, update); err != nil {
		return err
	}

	return nil
}

func createUpdateFilterBlueprint(filter *models.Filter, currentTime time.Time) bson.M {

	update := bson.M{
		"$set": bson.M{
			"filter.events":      filter.Events,
			"filter.instance_id": filter.InstanceID,
		},
		"$setOnInsert": bson.M{
			"last_updated": currentTime,
		},
	}

	return update
}

func createUpdateFilterOutput(filter *models.Filter, currentTime time.Time) bson.M {

	var downloads models.Downloads

	var update bson.M

	if filter.Downloads.XLS.URL != "" {
		downloads.XLS = filter.Downloads.XLS
	}

	if filter.Downloads.CSV.URL != "" {
		downloads.CSV = filter.Downloads.CSV
	}

	if filter.Downloads.JSON.URL != "" {
		downloads.JSON = filter.Downloads.JSON
	}

	// Don't bother checking for JSON as it doesn't get generated at the moment
	if downloads.CSV.URL != "" && downloads.XLS.URL != "" {
		update = bson.M{
			"$set": bson.M{
				"downloads": downloads,
				"events":    filter.Events,
				"state":     completed,
			},
			"$setOnInsert": bson.M{
				"last_updated": currentTime,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				"downloads": downloads,
				"events":    filter.Events,
			},
			"$setOnInsert": bson.M{
				"last_updated": currentTime,
			},
		}
	}

	return update
}

func (s *FilterStore) checkFilterState(filterID string) error {
	filter, err := s.GetFilter(filterID)
	if err != nil {
		return err
	}

	if filter.State == submitted {
		return errForbidden
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
