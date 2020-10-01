package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-graph/v2/graph/driver"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/graph_driver.go -pkg mock . GraphDriver
//go:generate moq -out mock/mongo.go -pkg mock . MongoDB

// GraphDriver in an alias to the graph Driver interface
type GraphDriver interface {
	driver.Driver
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// MongoDB defines the required methods from MongoDB package
type MongoDB interface {
	api.DataStore
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Close(ctx context.Context) error
}
