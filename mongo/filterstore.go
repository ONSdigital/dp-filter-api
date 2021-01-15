package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongolib "github.com/ONSdigital/dp-mongodb"
	mongohealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FilterStore containing all filter jobs stored in mongodb
type FilterStore struct {
	Session           *mgo.Session
	host              string
	db                string
	filtersCollection string
	outputsCollection string
	healthCheckClient *mongohealth.CheckMongoClient
}

// CreateFilterStore which can store, update and fetch filter jobs
func CreateFilterStore(cfg config.MongoConfig, host string) (*FilterStore, error) {
	session, err := mgo.Dial(cfg.BindAddr)
	if err != nil {
		return nil, err
	}
	filterStore := &FilterStore{
		Session:           session,
		host:              host,
		db:                cfg.Database,
		filtersCollection: cfg.FiltersCollection,
		outputsCollection: cfg.OutputsCollection,
	}

	databaseCollectionBuilder := make(map[mongohealth.Database][]mongohealth.Collection)
	databaseCollectionBuilder[(mongohealth.Database)(cfg.Database)] = []mongohealth.Collection{(mongohealth.Collection)(cfg.FiltersCollection), (mongohealth.Collection)(cfg.OutputsCollection)}

	client := mongohealth.NewClientWithCollections(session, databaseCollectionBuilder)
	filterStore.healthCheckClient = &mongohealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}
	return filterStore, nil
}

// AddFilter to the data store.
func (s *FilterStore) AddFilter(filter *models.Filter) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	// Initialise with a timestamp
	var err error
	filter.UniqueTimestamp, err = bson.NewMongoTimestamp(time.Now(), 1)
	if err != nil {
		return nil, err
	}

	// set eTag value to current hash of the filter
	filter.ETag, err = filter.Hash()
	if err != nil {
		return nil, err
	}

	// Insert filter to database
	if err = session.DB(s.db).C(s.filtersCollection).Insert(filter); err != nil {
		return nil, err
	}

	validateFilter(filter)
	return filter, nil
}

// GetFilter returns a single filter, if not found return an error.
// An optional eTag can be provided to assure that a filter has not been modified since the hash was generated
func (s *FilterStore) GetFilter(filterID, eTagSelector string) (*models.Filter, error) {
	session := s.Session.Copy()
	defer session.Close()

	return s.getFilterWithSession(session, filterID, 0, eTagSelector)
}

// get a filter with the provided session.
// Optional timestamp and eTag can be provided to assure that a filter has not been modified since expected.
func (s *FilterStore) getFilterWithSession(session *mgo.Session, filterID string, timestamp bson.MongoTimestamp, eTagSelector string) (*models.Filter, error) {

	query := selector(filterID, "", timestamp, eTagSelector)

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

// UpdateFilter replaces the stored filter properties.
func (s *FilterStore) UpdateFilter(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// calculate the new eTag hash for the filter that would result from applying the update
	newETag, err = s.getNewETagForUpdate(session, timestamp, eTagSelector, currentFilter, updatedFilter)
	if err != nil {
		return "", err
	}

	// create selector query
	selector := selector(updatedFilter.FilterID, "", timestamp, eTagSelector)

	// create update query from updatedFilter and newly generated eTag
	update, err := mongolib.WithUpdates(bson.M{
		"$set": bson.M{
			"events":      updatedFilter.Events,
			"instance_id": updatedFilter.InstanceID,
			"published":   updatedFilter.Published,
			"e_tag":       newETag,
		},
	})
	if err != nil {
		return "", err
	}

	// execute the update against MongoDB to atomically check and update the filter
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// GetFilterDimension return a single dimension, along with the filter eTag hash
func (s *FilterStore) GetFilterDimension(filterID string, name, eTagSelector string) (dimension *models.Dimension, eTag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// create selector query
	selector := selector(filterID, name, 0, eTagSelector)
	dimensionSelect := bson.M{"dimensions.$": 1}

	var result models.Filter
	if err := session.DB(s.db).C(s.filtersCollection).Find(selector).Select(dimensionSelect).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, "", filters.ErrDimensionNotFound
		}
		return nil, "", err
	}

	return &result.Dimensions[0], result.ETag, nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(filterID, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
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

	// define selector query
	selector := selector(filterID, "", timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = s.getNewETagForAddDimensions(session, timestamp, eTagSelector, currentFilter, filterID, list)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := mongolib.WithUpdates(bson.M{
		"$set": bson.M{"dimensions": list, "e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// run the query
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(filterID, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// define selector query
	selector := selector(filterID, "", timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = s.getNewETagForRemoveDimension(session, timestamp, eTagSelector, currentFilter, filterID, name)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := mongolib.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions": bson.M{"name": name}},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(selector, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	if info.Updated == 0 {
		return "", filters.ErrDimensionNotFound
	}

	return newETag, nil
}

// AddFilterDimensionOption to a filter.
func (s *FilterStore) AddFilterDimensionOption(filterID, name, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	return s.AddFilterDimensionOptions(filterID, name, []string{option}, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptions adds the provided options to a filter. The number of successfully added options is returned, along with an error.
func (s *FilterStore) AddFilterDimensionOptions(filterID, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = s.getNewETagForAddDimensionOptions(session, timestamp, eTagSelector, currentFilter, filterID, name, options)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := mongolib.WithUpdates(bson.M{
		"$addToSet": bson.M{"dimensions.$.options": bson.M{"$each": options}},
		"$set":      bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute update
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = s.getNewETagForRemoveDimensionOptions(session, timestamp, eTagSelector, currentFilter, filterID, name, []string{option})
	if err != nil {
		return "", err
	}

	// define update query
	update, err := mongolib.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions.$.options": option},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	info, err := session.DB(s.db).C(s.filtersCollection).UpdateAll(selector, update)
	if err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	// document was match but nothing was removed
	if info.Updated == 0 {
		return "", filters.ErrDimensionOptionNotFound
	}

	return newETag, nil
}

// RemoveFilterDimensionOptions removes the provided options from a filter. If an error happens, it is returned.
func (s *FilterStore) RemoveFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	session := s.Session.Copy()
	defer session.Close()

	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = s.getNewETagForRemoveDimensionOptions(session, timestamp, eTagSelector, currentFilter, filterID, name, options)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := mongolib.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions.$.options": bson.M{"$in": options}},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	if err := session.DB(s.db).C(s.filtersCollection).Update(selector, update); err != nil {
		if err == mgo.ErrNotFound {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// selector creates a select query for mongoDB with the provided parameters
// - filterID represents the ID of the filter document that we want to query. It is compulsory.
// - dimensionName is the name of a dimension that needs to be matched. It is optional.
// - timestamp is a unique MongoDB timestamp to be matched to prevent race conditions. It is optional.
// - eTagselector is a unique hash of a filter document to be matched to prevent race conditions. It is optional.
func selector(filterID, dimensionName string, timestamp bson.MongoTimestamp, eTagSelector string) bson.M {
	selector := bson.M{"filter_id": filterID}
	if dimensionName != "" {
		selector["dimensions"] = bson.M{"$elemMatch": bson.M{"name": dimensionName}}
	}
	if timestamp > 0 {
		selector["unique_timestamp"] = timestamp
	}
	if eTagSelector != "" {
		selector["e_tag"] = eTagSelector
	}
	return selector
}

// CreateFilterOutput creates a filter output resource
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
func (s *FilterStore) UpdateFilterOutput(filter *models.Filter, timestamp bson.MongoTimestamp, eTag string) error {
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

// AddEventToFilterOutput adds the given event to the filter output of the given ID
func (s *FilterStore) AddEventToFilterOutput(filterOutputID string, event *models.Event) error {
	session := s.Session.Copy()
	defer session.Close()

	info, err := session.DB(s.db).C(s.outputsCollection).Upsert(bson.M{"filter_id": filterOutputID},
		bson.M{"$push": bson.M{"events": &event}, "$set": bson.M{"last_updated": time.Now().UTC()}})
	if err != nil {
		return err
	}

	if info.Updated == 0 {
		return filters.ErrFilterOutputNotFound
	}

	return nil
}

func createUpdateFilterOutput(filter *models.Filter) bson.M {

	var downloads models.Downloads
	state := models.CreatedState

	var update bson.M
	if filter.Downloads != nil {
		if filter.Downloads.XLS != nil {
			if filter.Downloads.XLS.HRef != "" || filter.Downloads.XLS.Skipped {
				downloads.XLS = filter.Downloads.XLS
			}
		}

		if filter.Downloads.CSV != nil {
			if filter.Downloads.CSV.HRef != "" || filter.Downloads.CSV.Skipped {
				downloads.CSV = filter.Downloads.CSV
			}
		}
	}

	if filter.State != "" {
		state = filter.State
	}

	updates := bson.M{
		"state":     state,
		"downloads": downloads,
		"published": filter.Published,
	}

	update = bson.M{
		"$set": updates,
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

// Checker calls the mongoDB healthcheck client Checker
func (s *FilterStore) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return s.healthCheckClient.Checker(ctx, state)
}

// Close closes the mongoDB session
func (s *FilterStore) Close(ctx context.Context) error {
	return mongolib.Close(ctx, s.Session)
}
