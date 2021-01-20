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
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	ShutdownTimeout            time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken        string        `envconfig:"DATASET_API_AUTH_TOKEN"           json:"-"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"               json:"-"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	EnablePrivateEndpoints     bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	DownloadServiceSecretKey   string        `envconfig:"DOWNLOAD_SERVICE_SECRET_KEY"      json:"-"`
	MaxRequestOptions          int           `envconfig:"MAX_REQUEST_OPTIONS"`
	MaxDatasetOptions          int           `envconfig:"MAX_DATASET_OPTIONS"`
	BatchMaxWorkers            int           `envconfig:"BATCH_MAX_WORKERS"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	MongoConfig                MongoConfig
}

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	BindAddr          string `envconfig:"MONGODB_BIND_ADDR"           json:"-"`
	Database          string `envconfig:"MONGODB_FILTERS_DATABASE"`
	FiltersCollection string `envconfig:"MONGODB_FILTERS_COLLECTION"`
	OutputsCollection string `envconfig:"MONGODB_OUTPUT_COLLECTION"`
	Limit             int    `envconfig:"MONGODB_LIMIT"`
	Offset            int    `envconfig:"MONGODB_OFFSET"`
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
		KafkaMaxBytes:              2000000,
		ShutdownTimeout:            5 * time.Second,
		DatasetAPIURL:              "http://localhost:22000",
		DatasetAPIAuthToken:        "FD0108EA-825D-411C-9B1D-41EF7727F465",
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		MaxRequestOptions:          1000, // Maximum number of options acceptable in an incoming Patch request. Compromise between one option per call (inefficient) and an order of 100k options per call, for census data (memory and computationally expensive)
		MaxDatasetOptions:          200,  // Maximum number of options requested to Dataset API in a single call.
		BatchMaxWorkers:            25,   // maximum number of concurrent go-routines requesting items concurrently from APIs with pagination
		MongoConfig: MongoConfig{
			BindAddr:          "localhost:27017",
			Database:          "filters",
			FiltersCollection: "filters",
			OutputsCollection: "filterOutputs",
			Limit:             100,
			Offset:            0,
		},
		ServiceAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ZebedeeURL:               "http://localhost:8082",
		EnablePrivateEndpoints:   true,
		DownloadServiceURL:       "http://localhost:23600",
		DownloadServiceSecretKey: "QB0108EZ-825D-412C-9B1D-41EF7747F462",
		KafkaVersion:             "1.0.2",
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
