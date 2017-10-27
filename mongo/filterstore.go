package mongo

import (
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Database variables
const (
	database          = "filters"
	filtersCollection = "filters"
)

// Filter job states
const (
	submitted = "submitted"
	completed = "completed"
)

// Error codes
var (
	errNotFound                  = errors.New("Not found")
	errNotAuthorised             = errors.New("Not authorised")
	errDimensionNotFound         = errors.New("Dimension not found")
	errFilterOrDimensionNotFound = errors.New("Bad request - filter or dimension not found")
	errForbidden                 = errors.New("Forbidden")
	errOptionNotFound            = errors.New("Option not found")
	errBadRequest                = errors.New("Bad request - filter job not found")
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

	filter.FilterID = uuid.NewV4().String()
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

	query := bson.M{"filter_job_id": filterID}
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
func (s *FilterStore) UpdateFilter(isAuthenticated bool, updatedFilter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	currentFilter, err := s.GetFilter(updatedFilter.FilterID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	if !isAuthenticated {
		if currentFilter.State == submitted {
			return errForbidden
		}
		// force updates to the downloads property to be blank, only authenticated user can do this
		updatedFilter.Downloads = models.Downloads{}
	}

	if updatedFilter.Downloads.XLS.URL == "" {
		updatedFilter.Downloads.XLS = currentFilter.Downloads.XLS
	}

	if updatedFilter.Downloads.CSV.URL == "" {
		updatedFilter.Downloads.CSV = currentFilter.Downloads.CSV
	}

	if updatedFilter.Downloads.JSON.URL == "" {
		updatedFilter.Downloads.JSON = currentFilter.Downloads.JSON
	}

	// Don't bother checking for JSON as it doesn't get generated at the moment
	if updatedFilter.Downloads.CSV.URL == "" && updatedFilter.Downloads.XLS.URL == "" {
		updatedFilter.State = completed
	}

	query := bson.M{"filter_job_id": updatedFilter.FilterID}
	if err = session.DB(database).C(filtersCollection).Update(query, &updatedFilter); err != nil {
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

	query := bson.M{"filter_job_id": filterID}
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

	queryFilter := bson.M{"filter_job_id": filterID}
	queryDimension := bson.M{"filter_job_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(queryFilter).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return errBadRequest
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
		return errBadRequest
	}

	for i, item := range list {
		if item.Name != dimension.Name {
			continue
		}
		replace := &models.Dimension{
			URL:     item.URL,
			Name:    dimension.Name,
			Options: dimension.Options,
		}
		list[i] = *replace
	}

	if list == nil {
		url := fmt.Sprintf("%s/filter/%s/dimensions/%s", s.host, dimension.FilterID, dimension.Name)
		d := models.Dimension{Name: dimension.Name, Options: dimension.Options, URL: url}
		list = append(list, d)
	}

	queryFilter := bson.M{"filter_job_id": dimension.FilterID}
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
			return errBadRequest
		}
		return err
	}

	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_job_id": filterID}
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
			return errBadRequest
		}
		return err
	}

	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_job_id": newOption.FilterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": newOption.Name}}}
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

	queryFilter := bson.M{"filter_job_id": filterID}
	var result models.Filter

	if err := session.DB(database).C(filtersCollection).Find(queryFilter).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errBadRequest
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

	queryDimension := bson.M{"filter_job_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
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
	if err := s.checkFilterState(filterID); err != nil {
		if err == errNotFound {
			return errBadRequest
		}
		return err
	}
	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_job_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update := bson.M{"$pull": bson.M{"dimensions.$.options": option}}

	info, err := session.DB(database).C(filtersCollection).UpdateAll(queryOptions, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	// No document was matched
	if info.Matched == 0 {
		return errBadRequest
	}

	// document was match but nothing was removed
	if info.Updated == 0 {
		return errNotFound
	}

	return nil
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
