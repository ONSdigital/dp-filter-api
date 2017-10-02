package api

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/models"
)

// DatasetAPIer - An interface used to access the DatasetAPI
type DatasetAPIer interface {
	GetInstance(context.Context, string) (*models.Instance, error)
}
