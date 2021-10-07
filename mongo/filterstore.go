package mongo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/v3/health"
	dpMongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	connectTimeoutInSeconds = 5
	queryTimeoutInSeconds   = 15
)

// FilterStore containing all filter jobs stored in mongodb
type FilterStore struct {
	URI               string
	Database          string
	FiltersCollection string
	OutputsCollection string
	Connection        *dpMongoDriver.MongoConnection
	Username          string
	Password          string
	CAFilePath        string
	healthCheckClient *dpMongoHealth.CheckMongoClient
}

func NewMongoTimestamp(t time.Time, c uint32) (int64, error) {
	u := t.Unix()
	if u < 0 || u > math.MaxUint32 {
		return -1, errors.New("invalid value for time")
	}

	i := int64(u<<32 | int64(c))

	return int64(i), nil
}

// CreateFilterStore which can store, update and fetch filter jobs
func CreateFilterStore(cfg config.MongoConfig, host string) (*FilterStore, error) {
	mongoConnection, err := dpMongoDriver.Open(&dpMongoDriver.MongoConnectionConfig{
		ConnectTimeoutInSeconds: connectTimeoutInSeconds,
		QueryTimeoutInSeconds:   queryTimeoutInSeconds,

		Username:                      cfg.Username,
		Password:                      cfg.Password,
		ClusterEndpoint:               cfg.BindAddr,
		Database:                      cfg.Database,
		IsSSL:                         cfg.IsSSL,
		IsStrongReadConcernEnabled:    cfg.EnableStrongReadConcern,
		IsWriteConcernMajorityEnabled: cfg.EnableMajorityWriteConcern,
	})

	if err != nil {
		return nil, err
	}

	filterStore := &FilterStore{
		Connection:        mongoConnection,
		URI:               host,
		Database:          cfg.Database,
		FiltersCollection: cfg.FiltersCollection,
		OutputsCollection: cfg.OutputsCollection,
	}

	databaseCollectionBuilder := make(map[dpMongoHealth.Database][]dpMongoHealth.Collection)
	databaseCollectionBuilder[(dpMongoHealth.Database)(cfg.Database)] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(cfg.FiltersCollection), (dpMongoHealth.Collection)(cfg.OutputsCollection)}

	client := dpMongoHealth.NewClientWithCollections(mongoConnection, databaseCollectionBuilder)
	filterStore.healthCheckClient = &dpMongoHealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}
	return filterStore, nil
}

// AddFilter to the data store.
func (s *FilterStore) AddFilter(ctx context.Context, filter *models.Filter) (*models.Filter, error) {
	// Initialise with a timestamp
	filter.UniqueTimestamp = primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1}

	var err error
	// set eTag value to current hash of the filter
	filter.ETag, err = filter.Hash(nil)
	if err != nil {
		return nil, err
	}

	// Insert filter to database
	if _, err = s.Connection.C(s.FiltersCollection).Insert(ctx, filter); err != nil {
		return nil, err
	}

	validateFilter(filter)
	return filter, nil
}

// GetFilter returns a single filter, if not found return an error.
// An optional eTag can be provided to assure that a filter has not been modified since the hash was generated
func (s *FilterStore) GetFilter(ctx context.Context, filterID, eTagSelector string) (*models.Filter, error) {
	return s.getFilterWithSession(ctx, s.Connection, filterID, primitive.Timestamp{}, eTagSelector)
}

// get a filter with the provided session.
// Optional timestamp and eTag can be provided to assure that a filter has not been modified since expected.
func (s *FilterStore) getFilterWithSession(ctx context.Context, connection *dpMongoDriver.MongoConnection, filterID string, timestamp primitive.Timestamp, eTagSelector string) (*models.Filter, error) {

	// ignore eTag for query, so that we can return the correct error if it does not match
	query := selector(filterID, "", timestamp, AnyETag)

	var result models.Filter

	if err := connection.C(s.FiltersCollection).FindOne(ctx, query, &result); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, filters.ErrFilterBlueprintNotFound
		}
		return nil, err
	}

	// If eTag was provided and did not match, return the corresponding error
	if eTagSelector != AnyETag && eTagSelector != result.ETag {
		return nil, filters.ErrFilterBlueprintConflict
	}

	validateFilter(&result)
	return &result, nil
}

// UpdateFilter replaces the stored filter properties.
func (s *FilterStore) UpdateFilter(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	// calculate the new eTag hash for the filter that would result from applying the update
	newETag, err = newETagForUpdate(currentFilter, updatedFilter)
	if err != nil {
		return "", err
	}

	// create selector query
	selector := selector(updatedFilter.FilterID, "", timestamp, eTagSelector)

	// create update query from updatedFilter and newly generated eTag
	update, err := dpMongoDriver.WithUpdates(bson.M{
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
	if _, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// GetFilterDimension return a single dimension, along with the filter eTag hash
func (s *FilterStore) GetFilterDimension(ctx context.Context, filterID string, name, eTagSelector string) (dimension *models.Dimension, err error) {
	// create selector query
	selector := selector(filterID, name, primitive.Timestamp{}, eTagSelector)
	dimensionSelect := bson.M{"dimensions.$": 1}

	var result models.Filter
	if err := s.Connection.C(s.FiltersCollection).Find(selector).Select(dimensionSelect).One(ctx, &result); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, filters.ErrDimensionNotFound
		}
		return nil, err
	}

	return &result.Dimensions[0], nil
}

// AddFilterDimension to a filter
func (s *FilterStore) AddFilterDimension(ctx context.Context, filterID, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	url := fmt.Sprintf("%s/filters/%s/dimensions/%s", s.URI, filterID, name)
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
	newETag, err = newETagForAddDimensions(currentFilter, filterID, list)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := dpMongoDriver.WithUpdates(bson.M{
		"$set": bson.M{"dimensions": list, "e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// run the query
	if _, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// RemoveFilterDimension from a filter
func (s *FilterStore) RemoveFilterDimension(ctx context.Context, filterID, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	// define selector query
	selector := selector(filterID, "", timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = newETagForRemoveDimension(currentFilter, filterID, name)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := dpMongoDriver.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions": bson.M{"name": name}},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	info, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update)
	if err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	if info.ModifiedCount == 0 {
		return "", filters.ErrDimensionNotFound
	}

	return newETag, nil
}

// AddFilterDimensionOption to a filter.
func (s *FilterStore) AddFilterDimensionOption(ctx context.Context, filterID, name, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	return s.AddFilterDimensionOptions(ctx, filterID, name, []string{option}, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptions adds the provided options to a filter. The number of successfully added options is returned, along with an error.
func (s *FilterStore) AddFilterDimensionOptions(ctx context.Context, filterID, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = newETagForAddDimensionOptions(currentFilter, filterID, name, options)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := dpMongoDriver.WithUpdates(bson.M{
		"$addToSet": bson.M{"dimensions.$.options": bson.M{"$each": options}},
		"$set":      bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute update
	if _, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// RemoveFilterDimensionOption from a filter
func (s *FilterStore) RemoveFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = newETagForRemoveDimensionOptions(currentFilter, filterID, name, []string{option})
	if err != nil {
		return "", err
	}

	// define update query
	update, err := dpMongoDriver.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions.$.options": option},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	info, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update)
	if err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	// document was match but nothing was removed
	if info.ModifiedCount == 0 {
		return "", filters.ErrDimensionOptionNotFound
	}

	return newETag, nil
}

// RemoveFilterDimensionOptions removes the provided options from a filter. If an error happens, it is returned.
func (s *FilterStore) RemoveFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error) {
	// define selector query
	selector := selector(filterID, name, timestamp, eTagSelector)

	// calculate the new eTag hash for the filter that would result from removing the dimension
	newETag, err = newETagForRemoveDimensionOptions(currentFilter, filterID, name, options)
	if err != nil {
		return "", err
	}

	// define update query
	update, err := dpMongoDriver.WithUpdates(bson.M{
		"$pull": bson.M{"dimensions.$.options": bson.M{"$in": options}},
		"$set":  bson.M{"e_tag": newETag},
	})
	if err != nil {
		return "", err
	}

	// execute the query
	if _, err := s.Connection.C(s.FiltersCollection).Update(ctx, selector, update); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return "", filters.ErrFilterBlueprintConflict
		}
		return "", err
	}

	return newETag, nil
}

// selector creates a select query for mongoDB with the provided parameters
// - filterID represents the ID of the filter document that we want to query. Required.
// - dimensionName is the name of a dimension that needs to be matched. Optional.
// - timestamp is a unique MongoDB timestamp to be matched to prevent race conditions. Optional.
// - eTagselector is a unique hash of a filter document to be matched to prevent race conditions. Optional.
func selector(filterID, dimensionName string, timestamp primitive.Timestamp, eTagSelector string) bson.M {
	selector := bson.M{"filter_id": filterID}
	if dimensionName != "" {
		selector["dimensions"] = bson.M{"$elemMatch": bson.M{"name": dimensionName}}
	}
	if !timestamp.IsZero() {
		selector["unique_timestamp"] = timestamp
	}
	if eTagSelector != AnyETag {
		selector["e_tag"] = eTagSelector
	}
	return selector
}

// CreateFilterOutput creates a filter output resource
func (s *FilterStore) CreateFilterOutput(ctx context.Context, filter *models.Filter) (err error) {
	filter.UniqueTimestamp = primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1}
	if err != nil {
		return
	}

	_, err = s.Connection.C(s.OutputsCollection).Insert(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

// GetFilterOutput returns a filter output resource
func (s *FilterStore) GetFilterOutput(ctx context.Context, filterID string) (*models.Filter, error) {
	query := bson.M{"filter_id": filterID}
	var result *models.Filter

	if err := s.Connection.C(s.OutputsCollection).FindOne(ctx, query, &result); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, filters.ErrFilterOutputNotFound
		}

		return nil, err
	}

	return result, nil
}

// UpdateFilterOutput updates a filter output resource
func (s *FilterStore) UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error {
	update, err := dpMongoDriver.WithUpdates(createUpdateFilterOutput(filter))
	if err != nil {
		return err
	}

	if _, err = s.Connection.C(s.OutputsCollection).
		Update(ctx, bson.M{"filter_id": filter.FilterID, "unique_timestamp": timestamp}, update); err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return filters.ErrFilterOutputConflict
		}
	}

	return err
}

// AddEventToFilterOutput adds the given event to the filter output of the given ID
func (s *FilterStore) AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error {
	info, err := s.Connection.C(s.OutputsCollection).Upsert(ctx, bson.M{"filter_id": filterOutputID},
		bson.M{"$push": bson.M{"events": &event}, "$set": bson.M{"last_updated": time.Now().UTC()}})
	if err != nil {
		return err
	}

	if info.UpsertedCount == 0 && info.ModifiedCount == 0 {
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
	return s.Connection.Close(ctx)
}
