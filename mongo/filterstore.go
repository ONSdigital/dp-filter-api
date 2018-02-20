package mongo

import (
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	errFilterBadRequest          = errors.New("Bad request - filter blueprint not found")
)

// FilterStore containing all filter jobs stored in mongodb
type FilterStore struct {
	Session           *mgo.Session
	host              string
	db                string
	filtersCollection string
	outputsCollection string
}

// CreateFilterStore which can store, update and fetch filter jobs
func CreateFilterStore(cfg config.MongoConfig, host string) (*FilterStore, error) {
	session, err := mgo.Dial(cfg.BindAddr)
	if err != nil {
		return nil, err
	}
	return &FilterStore{
		Session:           session,
		host:              host,
		db:                cfg.Database,
		filtersCollection: cfg.FiltersCollection,
		outputsCollection: cfg.OutputsCollection,
	}, nil
}

// AddFilter to the data store
func (s *FilterStore) AddFilter(host string, filter *models.Filter) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	if err := session.DB(s.db).C(s.filtersCollection).Insert(filter); err != nil {
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

	if err := session.DB(s.db).C(s.filtersCollection).Find(query).One(&result); err != nil {
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
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	return nil
}

// GetFilterDimension return a single dimension
func (s *FilterStore) GetFilterDimension(filterID string, name string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryDimension := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	dimensionSelect := bson.M{"dimensions": 1}
	var result models.Filter

	if err := session.DB(s.db).C(s.filtersCollection).Find(queryDimension).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return errDimensionNotFound
		}
		return err
	}

	return nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension) error {
	session := s.Session.Copy()
	defer session.Close()

	url := fmt.Sprintf("%s/filters/%s/dimensions/%s", s.host, filterID, name)
	d := models.Dimension{Name: name, Options: options, URL: url}

	list := dimensions
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

	queryFilter := bson.M{"filter_id": filterID}
	update := bson.M{"$set": bson.M{"dimensions": list}}

	if err := session.DB(s.db).C(s.filtersCollection).Update(queryFilter, update); err != nil {
		if err == mgo.ErrNotFound {
			return errNotFound
		}
		return err
	}

	return nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID, name string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_id": filterID}
	update := bson.M{"$pull": bson.M{"dimensions": bson.M{"name": name}}}

	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(queryFilter, update)
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
func (s *FilterStore) AddFilterDimensionOption(filterID, name, option string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update := bson.M{"$addToSet": bson.M{"dimensions.$.options": option}}

	if err := session.DB(s.db).C(s.filtersCollection).Update(queryOptions, update); err != nil {
		if err == mgo.ErrNotFound {
			return errDimensionNotFound
		}
		return err
	}

	return nil
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string) error {
	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": filterID, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update := bson.M{"$pull": bson.M{"dimensions.$.options": option}}

	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(queryOptions, update)
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

// CreateFilterOutput creates a filter ouput resource
func (s *FilterStore) CreateFilterOutput(filter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	if err := session.DB(s.db).C(s.outputsCollection).Insert(filter); err != nil {
		return err
	}

	return nil
}

// GetFilterOutput returns a filter output resource
func (s *FilterStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	query := bson.M{"filter_id": filterID}
	var result *models.Filter

	if err := session.DB(s.db).C(s.outputsCollection).Find(query).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errFilterOutputNotFound
		}

		return nil, err
	}

	return result, nil
}

// UpdateFilterOutput updates a filter output resource
func (s *FilterStore) UpdateFilterOutput(filter *models.Filter) error {
	session := s.Session.Copy()
	defer session.Close()

	update := createUpdateFilterOutput(filter, time.Now())

	if err := session.DB(s.db).C(s.outputsCollection).
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
			"filter.published":   filter.Published,
		},
		"$setOnInsert": bson.M{
			"last_updated": currentTime,
		},
	}

	return update
}

func createUpdateFilterOutput(filter *models.Filter, currentTime time.Time) bson.M {
	var downloads models.Downloads
	state := models.CreatedState
	var update bson.M
	if filter.Downloads != nil {
		if filter.Downloads.XLS.URL != "" {
			downloads.XLS = filter.Downloads.XLS
		}

		if filter.Downloads.CSV.URL != "" {
			downloads.CSV = filter.Downloads.CSV
		}

		if filter.Downloads.JSON.URL != "" {
			downloads.JSON = filter.Downloads.JSON
		}
	}

	if filter.State != "" {
		state = filter.State
	}

	// Don't bother checking for JSON as it doesn't get generated at the moment
	if downloads.CSV.URL != "" && downloads.XLS.URL != "" {
		update = bson.M{
			"$set": bson.M{
				"downloads": downloads,
				"events":    filter.Events,
				"state":     models.CompletedState,
				"published": filter.Published,
			},
			"$setOnInsert": bson.M{
				"last_updated": currentTime,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				"state":     state,
				"downloads": downloads,
				"events":    filter.Events,
				"published": filter.Published,
			},
			"$setOnInsert": bson.M{
				"last_updated": currentTime,
			},
		}
	}

	return update
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
