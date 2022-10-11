package config

import (
	"fmt"
	"time"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	FilterOutputSubmittedTopic string        `envconfig:"FILTER_JOB_SUBMITTED_TOPIC"`
	Host                       string        `envconfig:"HOST"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"             json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ShutdownTimeout            time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
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
	DefaultMaxLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	AssertDatasetType          bool          `envconfig:"ASSERT_DATASET_TYPE"`
	FilterFlexAPIURL           string        `envconfig:"FILTER_FLEX_API_URL"`
	EnableFilterOutputs        bool          `envconfig:"ENABLE_FILTER_OUTPUTS_CHECK"`
	FilterOutputToken          bool          `envconfig:"FILTER_OUTPUT_TOKEN"`

	MongoConfig
}

type MongoConfig struct {
	mongodriver.MongoDriverConfig

	Limit  int `envconfig:"MONGODB_LIMIT"`
	Offset int `envconfig:"MONGODB_OFFSET"`
}

var cfg *Config

const (
	FiltersCollection = "FiltersCollection"
	OutputsCollection = "OutputsCollection"
)

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		Host:                       "http://localhost:22100",
		BindAddr:                   ":22100",
		Brokers:                    []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:               "1.0.2",
		FilterOutputSubmittedTopic: "filter-job-submitted",
		KafkaMaxBytes:              2000000,
		ShutdownTimeout:            5 * time.Second,
		DatasetAPIURL:              "http://localhost:22000",
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		MaxRequestOptions:          1000, // Maximum number of options acceptable in an incoming Patch request. Compromise between one option per call (inefficient) and an order of 100k options per call, for census data (memory and computationally expensive)
		MaxDatasetOptions:          200,  // Maximum number of options requested to Dataset API in a single call by a list of ids
		BatchMaxWorkers:            25,   // Maximum number of concurrent go-routines requesting items concurrently from APIs with pagination
		DefaultMaxLimit:            1000, // Maximum limit allowed for paginated calls
		ServiceAuthToken:           "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ZebedeeURL:                 "http://localhost:8082",
		EnablePrivateEndpoints:     true,
		DownloadServiceURL:         "http://localhost:23600",
		DownloadServiceSecretKey:   "QB0108EZ-825D-412C-9B1D-41EF7747F462",
		AssertDatasetType:          false,
		FilterFlexAPIURL:           "http://localhost:27100",
		MongoConfig: MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "filters",
				Collections:                   map[string]string{FiltersCollection: "filters", OutputsCollection: "filterOutputs"},
				ReplicaSet:                    "",
				IsStrongReadConcernEnabled:    false,
				IsWriteConcernMajorityEnabled: true,
				ConnectTimeout:                5 * time.Second,
				QueryTimeout:                  15 * time.Second,
				TLSConnectionConfig: mongodriver.TLSConnectionConfig{
					IsSSL: false,
				},
			},
			Limit:  20, // Default limit for mongoDB queries that do not provide an explicit limit
			Offset: 0,  // Default offset for mongoDB queries that do not provide an explicit offset
		},
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed tp pas: %w", err)
	}

	return cfg, nil
}
