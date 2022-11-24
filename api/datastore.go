package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/mongo"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

//go:generate moq -out mock/datastore.go -pkg mock . DataStore

// DataStore - A interface used to store filters
type DataStore interface {
	AddFilter(ctx context.Context, filter *models.Filter) (*models.Filter, error)
	GetFilter(ctx context.Context, filterID, eTagSelector string) (*models.Filter, error)
	UpdateFilter(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	GetFilterDimension(ctx context.Context, filterID string, name, eTagSelector string) (dimension *models.Dimension, err error)
	AddFilterDimension(ctx context.Context, filterID, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimension(ctx context.Context, filterID, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOption(ctx context.Context, filterID, name, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	AddFilterDimensionOptions(ctx context.Context, filterID, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	RemoveFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (newETag string, err error)
	CreateFilterOutput(ctx context.Context, filter *models.Filter) error
	GetFilterOutput(ctx context.Context, filterOutputID string) (*models.Filter, error)
	UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error
	AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error
	RunTransaction(ctx context.Context, retry bool, fn mongodriver.TransactionFunc) (interface{}, error)
}

func (api *FilterAPI) testTrans(w http.ResponseWriter, r *http.Request) {
	type simpleObject struct {
		ID    int    `bson:"_id"`
		State string `bson:"state"`
	}

	// Create the collection and insert the object outside the transaction
	var (
		ctx    = r.Context()
		obj    = simpleObject{ID: 1, State: "1"}
		result interface{}
		err    error
	)
	_, err = api.dataStore.(*mongo.FilterStore).Connection.Collection("test-collection").Upsert(ctx, bson.M{"_id": obj.ID}, bson.M{"$set": obj})
	if err != nil {
		setErrorCodeFromError(w, err)
		return
	}

	result, err = api.dataStore.RunTransaction(r.Context(), false, func(transactionCtx context.Context) (interface{}, error) {
		for i := 1; i < 50; i++ {
			err = api.dataStore.(*mongo.FilterStore).Connection.Collection("test-collection").FindOne(transactionCtx, bson.M{"_id": obj.ID}, &obj)
			if obj.State != strconv.Itoa(i) {
				return nil, errors.New("object in incorrect state")
			}

			time.Sleep(1 * time.Second)

			obj.State = strconv.Itoa(i + 1)
			_, err = api.dataStore.(*mongo.FilterStore).Connection.Collection("test-collection").Update(transactionCtx, bson.M{"_id": obj.ID}, bson.M{"$set": obj})
			if err != nil {
				return nil, err
			}
		}

		return obj, nil
	})
	if err != nil {
		setErrorCodeFromError(w, err)
		return
	}

	setJSONContentType(w)
	err = WriteJSONBody(ctx, result, w, nil)
	if err != nil {
		setErrorCode(w, err)
		return
	}
}
