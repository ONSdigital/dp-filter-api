package mongo

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// getNewETag generates a new eTag that corresponds to the updated filter
// - currentFilter (optional). If provided, it will be used to update the eTag hash. Otherwise, the filter will be read from the DB so that we can update the hash.
// - eTagSelector (optional). If provided and no currentFilter was provided, it will be used to ensure we read the expected filter from the DB.
func (s *FilterStore) getNewETag(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, filterID string, updater func(models.Filter) models.Filter) (eTag string, err error) {
	// get current filter, if it is not provided
	if currentFilter == nil {
		currentFilter, err = s.getFilterWithSession(session, filterID, timestamp, eTagSelector)
		if err != nil {
			return "", err
		}
	}

	// execute the update, to have the filter that would result after a successful update
	f := updater(*currentFilter)

	// calculate the new eTag hash for the filter that would result from applying the update
	return f.Hash()
}

func (s *FilterStore) getNewETagForUpdate(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, update *models.Filter) (eTag string, err error) {
	return s.getNewETag(session, timestamp, eTagSelector, currentFilter, update.FilterID, func(filter models.Filter) models.Filter {
		filter.Events = update.Events
		filter.InstanceID = update.InstanceID
		filter.Published = update.Published
		return filter
	})
}

func (s *FilterStore) getNewETagForAddDimensions(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, filterID string, dims []models.Dimension) (eTag string, err error) {
	return s.getNewETag(session, timestamp, eTagSelector, currentFilter, filterID, func(filter models.Filter) models.Filter {
		filter.Dimensions = append(filter.Dimensions, dims...)
		return filter
	})
}

func (s *FilterStore) getNewETagForRemoveDimension(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, filterID, dimensionNameToRemove string) (eTag string, err error) {
	return s.getNewETag(session, timestamp, eTagSelector, currentFilter, filterID, func(filter models.Filter) models.Filter {
		updatedDimensions := []models.Dimension{}
		for _, dim := range filter.Dimensions {
			if dim.Name != dimensionNameToRemove {
				updatedDimensions = append(updatedDimensions, dim)
			}
		}
		filter.Dimensions = updatedDimensions
		return filter
	})
}

func (s *FilterStore) getNewETagForAddDimensionOptions(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, filterID, dimensionName string, options []string) (eTag string, err error) {
	return s.getNewETag(session, timestamp, eTagSelector, currentFilter, filterID, func(filter models.Filter) models.Filter {
		updatedDimensions := []models.Dimension{}
		for _, dim := range filter.Dimensions {
			if dim.Name == dimensionName {
				dim.Options = utils.CreateArray(utils.CreateMap(append(dim.Options, options...)))
			}
			updatedDimensions = append(updatedDimensions, dim)
		}
		filter.Dimensions = updatedDimensions
		return filter
	})
}

func (s *FilterStore) getNewETagForRemoveDimensionOptions(session *mgo.Session, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter, filterID, dimensionName string, options []string) (eTag string, err error) {
	return s.getNewETag(session, timestamp, eTagSelector, currentFilter, filterID, func(filter models.Filter) models.Filter {
		rmOptions := utils.CreateMap(options)
		for _, dim := range filter.Dimensions {
			if dim.Name == dimensionName {
				opts := []string{}
				for _, opt := range dim.Options {
					if _, found := rmOptions[opt]; !found {
						opts = append(opts, opt)
					}
				}
				dim.Options = opts
				break
			}
		}
		return filter
	})
}
