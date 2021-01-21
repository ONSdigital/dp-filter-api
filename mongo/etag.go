package mongo

import (
	"fmt"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/globalsign/mgo/bson"
)

func newETagForUpdate(currentFilter *models.Filter, update *models.Filter) (eTag string, err error) {
	b, err := bson.Marshal(update)
	if err != nil {
		return "", err
	}
	return currentFilter.Hash(b)
}

func newETagForAddDimensions(currentFilter *models.Filter, filterID string, dims []models.Dimension) (eTag string, err error) {
	b, err := bson.Marshal(dims)
	if err != nil {
		return "", err
	}
	return currentFilter.Hash(b)
}

func newETagForRemoveDimension(currentFilter *models.Filter, filterID, dimensionNameToRemove string) (eTag string, err error) {
	b := []byte(fmt.Sprintf("RemoveDimension %s", dimensionNameToRemove))
	return currentFilter.Hash(b)
}

func newETagForAddDimensionOptions(currentFilter *models.Filter, filterID, dimensionName string, options []string) (eTag string, err error) {
	b := []byte(fmt.Sprintf("AddDimensionOptions %v", options))
	return currentFilter.Hash(b)
}

func newETagForRemoveDimensionOptions(currentFilter *models.Filter, filterID, dimensionName string, options []string) (eTag string, err error) {
	b := []byte(fmt.Sprintf("RemoveDimensionOptions %v", options))
	return currentFilter.Hash(b)
}
