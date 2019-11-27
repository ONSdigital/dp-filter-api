package initialise

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/go-ns/kafka"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	AuditProducer                 bool
	FilterOutputSubmittedProducer bool
	FilterStore                   bool
	ObservationStore              bool
}

const (
	// AuditProducer represents a name for
	// the producer that writes to an audit topic
	AuditProducer = "audit-producer"
	// FilterOutputSubmittedProducer represents a name for
	// the producer that writes to a filter output submitted topic
	FilterOutputSubmittedProducer = "filter-output-submitted-producer"
)

// GetFilterStore returns an initialised connection to filter store (mongo database)
func (e *ExternalServiceList) GetFilterStore(cfg *config.Config) (dataStore *mongo.FilterStore, err error) {
	dataStore, err = mongo.CreateFilterStore(cfg.MongoConfig, cfg.Host)
	if err == nil {
		e.FilterStore = true
	}

	return
}

// GetObservationStore returns an initialised connection to observation store (graph database)
func (e *ExternalServiceList) GetObservationStore() (observationStore *graph.DB, err error) {
	observationStore, err = graph.New(context.Background(), graph.Subsets{Observation: true})
	if err == nil {
		e.ObservationStore = true
	}

	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic, name string, envMax int) (kafkaProducer kafka.Producer, err error) {
	kafkaProducer, err = kafka.NewProducer(kafkaBrokers, topic, envMax)
	if err == nil {
		switch {
		case name == AuditProducer:
			e.AuditProducer = true
		case name == FilterOutputSubmittedProducer:
			e.FilterOutputSubmittedProducer = true
		}
	}

	return
}
