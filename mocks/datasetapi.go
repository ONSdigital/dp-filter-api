package mocks

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/rchttp"
)

type DatasetAPI struct {
}

func NewDatasetAPI(client *rchttp.Client, url, token string) *DatasetAPI {
	return &DatasetAPI{}
}

func (ds *DatasetAPI) GetInstance(ctx context.Context, id string) (*models.Instance, error) {
	return &models.Instance{}, nil
}
