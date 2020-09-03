package initialise

import (
	"context"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/mongo"
	"github.com/ONSdigital/dp-graph/v2/graph"
	kafka "github.com/ONSdigital/dp-kafka"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	FilterOutputSubmittedProducer bool
	FilterStore                   bool
	ObservationStore              bool
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
func (e *ExternalServiceList) GetProducer(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafkaProducer *kafka.Producer, err error) {
	producerChannels := kafka.CreateProducerChannels()
	kafkaProducer, err = kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, producerChannels)
	if err != nil {
		return
	}
	e.FilterOutputSubmittedProducer = true
	return
}
