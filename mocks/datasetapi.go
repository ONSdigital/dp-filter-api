package mocks

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/rchttp"
)

type DatasetAPI struct {
	InstanceNotFound    bool
	InternalServerError bool
}

var (
	instanceNotFoundError = errors.New("Instance not found")
)

func NewDatasetAPI(client *rchttp.Client, url, token string) *DatasetAPI {
	return &DatasetAPI{}
}

func (ds *DatasetAPI) GetInstance(ctx context.Context, id string) (*models.Instance, error) {
	if ds.InternalServerError {
		return nil, internalServerError
	}

	if ds.InstanceNotFound {
		return nil, instanceNotFoundError
	}

	return &models.Instance{}, nil
}
