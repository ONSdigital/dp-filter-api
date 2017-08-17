package config

import "github.com/ian-kent/gofigure"

// Config is the filing resource handler config
type Config struct {
	BindAddr                string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	Brokers                 []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The Kafka broker addresses"`
	FilterJobSubmittedTopic string   `env:"FILTER_JOB_SUBMITTED_TOPIC" flag:"filter-job-submitted-topic" flagDesc:"The Kafka topic to write submitted filter job messages to"`
	Host                    string   `env:"HOST" flag:"host" flagDesc:"The host name used to build URLs"`
	KafkaMaxBytes           string   `env:"KAFKA_MAX_BYTES" flag:"kafka-max-bytes" flagDesc:"The maximum permitted size of a message. Should be set equal to or smaller than the broker's 'message.max.bytes'"`
	PostgresURL             string   `env:"POSTGRES_URL" flag:"postgres-url" flagDesc:"The URL address to connect to a postgres instance'"`
	SecretKey               string   `env:"SECRET_KEY" flag:"secret-key" flagDesc:"A secret key used authentication"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                ":22100",
		Brokers:                 []string{"localhost:9092"},
		FilterJobSubmittedTopic: "filter-job-submitted",
		Host:          "http://localhost:22100",
		KafkaMaxBytes: "2000000",
		PostgresURL:   "user=dp dbname=FilterJobs sslmode=disable",
		SecretKey:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
	}

	if err := gofigure.Gofigure(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
