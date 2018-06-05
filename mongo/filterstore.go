package mongo

import (
	"fmt"
	"time"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	mongolib "github.com/ONSdigital/go-ns/mongo"
	"github.com/gedge/mgo"
	"github.com/gedge/mgo/bson"
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

	// Initialise with a timestamp
	var err error
	filter.UniqueTimestamp, err = bson.NewMongoTimestamp(time.Now(), 1)
	if err != nil {
		return nil, err
	}

	if err = session.DB(s.db).C(s.filtersCollection).Insert(filter); err != nil {
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
			return nil, filters.ErrFilterBlueprintNotFound
		}
		return nil, err
	}

	validateFilter(&result)
	return &result, nil
}

// UpdateFilter replaces the stored filter properties
func (s *FilterStore) UpdateFilter(updatedFilter *models.Filter, timestamp bson.MongoTimestamp) error {
	session := s.Session.Copy()
	defer session.Close()

	update, err := mongolib.WithUpdates(createUpdateFilterBlueprint(updatedFilter))
	if err != nil {
		return err
	}

	selector := bson.M{"filter_id": updatedFilter.FilterID, "unique_timestamp": timestamp}
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterBlueprintConflict
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
			return filters.ErrDimensionNotFound
		}
		return err
	}

	return nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp) error {
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

	selector := bson.M{"filter_id": filterID, "unique_timestamp": timestamp}
	update, err := mongolib.WithUpdates(bson.M{"$set": bson.M{"dimensions": list}})
	if err != nil {
		return err
	}

	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterBlueprintConflict
		}
		return err
	}

	return nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID, name string, timestamp bson.MongoTimestamp) error {
	session := s.Session.Copy()
	defer session.Close()

	queryFilter := bson.M{"filter_id": filterID, "unique_timestamp": timestamp}
	update, err := mongolib.WithUpdates(bson.M{"$pull": bson.M{"dimensions": bson.M{"name": name}}})
	if err != nil {
		return err
	}

	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(queryFilter, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterBlueprintConflict
		}
		return err
	}

	if info.Updated == 0 {
		return filters.ErrDimensionNotFound
	}

	return nil
}

// AddFilterDimensionOption to a filter
func (s *FilterStore) AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp) error {
	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": filterID, "unique_timestamp": timestamp, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update, err := mongolib.WithUpdates(bson.M{"$addToSet": bson.M{"dimensions.$.options": option}})
	if err != nil {
		return err
	}

	if err := session.DB(s.db).C(s.filtersCollection).Update(queryOptions, update); err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterBlueprintConflict
		}
		return err
	}

	return nil
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp) error {
	session := s.Session.Copy()
	defer session.Close()

	queryOptions := bson.M{"filter_id": filterID, "unique_timestamp": timestamp, "dimensions": bson.M{"$elemMatch": bson.M{"name": name}}}
	update, err := mongolib.WithUpdates(bson.M{"$pull": bson.M{"dimensions.$.options": option}})
	if err != nil {
		return err
	}

	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(queryOptions, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterBlueprintConflict
		}
		return err
	}

	// document was match but nothing was removed
	if info.Updated == 0 {
		return filters.ErrDimensionOptionNotFound
	}

	return nil
}

// CreateFilterOutput creates a filter ouput resource
func (s *FilterStore) CreateFilterOutput(filter *models.Filter) (err error) {
	session := s.Session.Copy()
	defer session.Close()

	filter.UniqueTimestamp, err = bson.NewMongoTimestamp(time.Now(), 1)
	if err != nil {
		return
	}

	return session.DB(s.db).C(s.outputsCollection).Insert(filter)
}

// GetFilterOutput returns a filter output resource
func (s *FilterStore) GetFilterOutput(filterID string) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	query := bson.M{"filter_id": filterID}
	var result *models.Filter

	if err := session.DB(s.db).C(s.outputsCollection).Find(query).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, filters.ErrFilterOutputNotFound
		}

		return nil, err
	}

	return result, nil
}

// UpdateFilterOutput updates a filter output resource
func (s *FilterStore) UpdateFilterOutput(filter *models.Filter, timestamp bson.MongoTimestamp) error {
	session := s.Session.Copy()
	defer session.Close()

	update, err := mongolib.WithUpdates(createUpdateFilterOutput(filter))
	if err != nil {
		return err
	}

	if err = session.DB(s.db).C(s.outputsCollection).
		Update(bson.M{"filter_id": filter.FilterID, "unique_timestamp": timestamp}, update); err != nil {
		if err == mgo.ErrNotFound {
			return filters.ErrFilterOutputConflict
		}
	}

	return err
}

func createUpdateFilterBlueprint(filter *models.Filter) bson.M {

	update := bson.M{
		"$set": bson.M{
			"filter.events":      filter.Events,
			"filter.instance_id": filter.InstanceID,
			"filter.published":   filter.Published,
		},
	}

	return update
}

func createUpdateFilterOutput(filter *models.Filter) bson.M {
	var downloads models.Downloads
	state := models.CreatedState
	var update bson.M
	if filter.Downloads != nil {
		if filter.Downloads.XLS != nil {
			if filter.Downloads.XLS.HRef != "" {
				downloads.XLS = filter.Downloads.XLS
			}
		}

		if filter.Downloads.CSV != nil {
			if filter.Downloads.CSV.HRef != "" {
				downloads.CSV = filter.Downloads.CSV
			}
		}
	}

	if filter.State != "" {
		state = filter.State
	}

	// Don't bother checking for JSON as it doesn't get generated at the moment
	if downloads.CSV != nil && downloads.CSV.HRef != "" && downloads.XLS != nil && downloads.XLS.HRef != "" {
		update = bson.M{
			"$set": bson.M{
				"downloads": downloads,
				"events":    filter.Events,
				"state":     models.CompletedState,
				"published": filter.Published,
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
		}
	}

	return update
}

func validateFilter(filter *models.Filter) {
	// Make sure all empty arrays are initialized
	if filter.Dimensions == nil {
		filter.Dimensions = []models.Dimension{}
	}

	if filter.Events == nil {
		filter.Events = []*models.Event{}
	}
}
