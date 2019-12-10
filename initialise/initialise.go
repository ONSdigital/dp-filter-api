package initialise

import (
	"context"
	"fmt"

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

// KafkaProducerName represents a type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafa Producers
const (
	Audit = iota
	FilterOutputSubmitted
)

var kafkaProducerNames = []string{"CSVExported", "Error"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// GetFilterStore returns an initialised connection to filter store (mongo database)
func (e *ExternalServiceList) GetFilterStore(cfg *config.Config) (dataStore *mongo.FilterStore, err error) {
	dataStore, err = mongo.CreateFilterStore(cfg.MongoConfig, cfg.Host)
	if err != nil {
		return
	}
	e.FilterStore = true

	return
}

// GetObservationStore returns an initialised connection to observation store (graph database)
func (e *ExternalServiceList) GetObservationStore() (observationStore *graph.DB, err error) {
	observationStore, err = graph.New(context.Background(), graph.Subsets{Observation: true})
	if err != nil {
		return
	}
	e.ObservationStore = true

	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic string, name KafkaProducerName, envMax int) (kafkaProducer kafka.Producer, err error) {
	kafkaProducer, err = kafka.NewProducer(kafkaBrokers, topic, envMax)
	if err == nil {
		switch {
		case name == Audit:
			e.AuditProducer = true
		case name == FilterOutputSubmitted:
			e.FilterOutputSubmittedProducer = true
		default:
			err = fmt.Errorf("Kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
		}
	}

	return
}
