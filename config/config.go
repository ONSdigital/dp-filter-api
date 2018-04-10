package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	FilterOutputSubmittedTopic string        `envconfig:"FILTER_JOB_SUBMITTED_TOPIC"`
	Host                       string        `envconfig:"HOST"`
	KafkaMaxBytes              string        `envconfig:"KAFKA_MAX_BYTES"`
	ShutdownTimeout            time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken        string        `envconfig:"DATASET_API_AUTH_TOKEN"      json:"-"`
	Neo4jURL                   string        `envconfig:"NEO4J_BIND_ADDR"             json:"-"`
	Neo4jPoolSize              int           `envconfig:"NEO4J_POOL_SIZE"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	MongoConfig                MongoConfig
	ServiceAuthToken           string `envconfig:"SERVICE_AUTH_TOKEN"          json:"-"`
	ZebedeeURL                 string `envconfig:"ZEBEDEE_URL"`
	EnablePrivateEndpoints     bool   `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
}

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	BindAddr          string `envconfig:"MONGODB_BIND_ADDR"           json:"-"`
	Database          string `envconfig:"MONGODB_FILTERS_DATABASE"`
	FiltersCollection string `envconfig:"MONGODB_FILTERS_COLLECTION"`
	OutputsCollection string `envconfig:"MONGODB_OUTPUT_COLLECTION"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		Host:                       "http://localhost:22100",
		BindAddr:                   ":22100",
		Brokers:                    []string{"localhost:9092"},
		FilterOutputSubmittedTopic: "filter-job-submitted",
		KafkaMaxBytes:              "2000000",
		ShutdownTimeout:            5 * time.Second,
		DatasetAPIURL:              "http://localhost:22000",
		DatasetAPIAuthToken:        "FD0108EA-825D-411C-9B1D-41EF7727F465",
		Neo4jURL:                   "bolt://localhost:7687",
		Neo4jPoolSize:              30,
		HealthCheckInterval:        30 * time.Second,
		MongoConfig: MongoConfig{
			BindAddr:          "localhost:27017",
			Database:          "filters",
			FiltersCollection: "filters",
			OutputsCollection: "filterOutputs",
		},
		ServiceAuthToken:       "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ZebedeeURL:             "http://localhost:8082",
		EnablePrivateEndpoints: false,
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	return cfg, nil
}
