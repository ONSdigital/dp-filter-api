package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	BindAddr                string        `envconfig:"BIND_ADDR"`
	Brokers                 []string      `envconfig:"KAFKA_ADDR"`
	FilterJobSubmittedTopic string        `envconfig:"FILTER_JOB_SUBMITTED_TOPIC"`
	Host                    string        `envconfig:"HOST"`
	KafkaMaxBytes           string        `envconfig:"KAFKA_MAX_BYTES"`
	PostgresURL             string        `envconfig:"POSTGRES_URL"`
	SecretKey               string        `envconfig:"SECRET_KEY"`
	ShutdownTimeout         time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	DatasetAPIURL           string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string        `envconfig:"DATASET_API_AUTH_TOKEN"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		Host:                    "http://localhost:22100",
		BindAddr:                ":22100",
		Brokers:                 []string{"localhost:9092"},
		FilterJobSubmittedTopic: "filter-job-submitted",
		KafkaMaxBytes:           "2000000",
		PostgresURL:             "user=dp dbname=FilterJobs sslmode=disable",
		SecretKey:               "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ShutdownTimeout:         5 * time.Second,
		DatasetAPIURL:           "http://localhost:22000",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
	}

	return cfg, envconfig.Process("", cfg)
}
