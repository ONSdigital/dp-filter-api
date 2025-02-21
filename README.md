# dp-filter-api

## Getting started

### mongo

* Run `brew install mongodb`
* Run `brew services start mongodb`

### Kafka

* Run `brew install zookeeper`
* Download kafka version 0.10.2.1
* Run `./kafka-server-start.sh ../config/server.properties`

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

### Configuration

* Run `brew install neo4j`
* Disable authentication in the config
* Run `brew services start neo4j`

| Environment variable         | Default                                                      | Description                                                                                                      |
| ---------------------------- |--------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :22100                                                       | The host and port to bind to                                                                                     |
| HOST                         | http://localhost:22100                                       | The host name used to build URLs                                                                                 |
| KAFKA_ADDR                   | localhost:9092                                               | The kafka broker addresses (can be comma separated)                                                              |
| KAFKA_MAX_BYTES              | 2000000                                                      | The maximum permitted size of a message. Should be set equal to or smaller than the broker's `message.max.bytes` |
| KAFKA_SEC_PROTO              | _unset_                                                      | if set to `TLS`, kafka connections will use TLS [1]                                                              |
| KAFKA_SEC_CLIENT_KEY         | _unset_                                                      | PEM for the client key [1]                                                                                       |
| KAFKA_SEC_CLIENT_CERT        | _unset_                                                      | PEM for the client certificate [1]                                                                               |
| KAFKA_SEC_CA_CERTS           | _unset_                                                      | CA cert chain for the server cert [1]                                                                            |
| KAFKA_SEC_SKIP_VERIFY        | false                                                        | ignores server certificate issues if `true` [1]                                                                  |
| FILTER_JOB_SUBMITTED_TOPIC   | filter-job-submitted                                         | The kafka topic to write messages to                                                                             |
| MONGODB_BIND_ADDR            | localhost:27017                                              | The MongoDB bind address                                                                                         |
| MONGODB_USERNAME             |                                                              | The MongoDB Username                                                                                             |
| MONGODB_PASSWORD             |                                                              | The MongoDB Password                                                                                             |
| MONGODB_DATABASE             | filters                                                      | The MongoDB database                                                                                             |
| MONGODB_COLLECTIONS          | FiltersCollection:filters, OutputsCollection:filterOutputs   | The MongoDB collections                                                                                          |
| MONGODB_REPLICA_SET          |                                                              | The name of the MongoDB replica set                                                                              |
| MONGODB_ENABLE_READ_CONCERN  | false                                                        | Switch to use (or not) majority read concern                                                                     |
| MONGODB_ENABLE_WRITE_CONCERN | true                                                         | Switch to use (or not) majority write concern                                                                    |
| MONGODB_CONNECT_TIMEOUT      | 5s                                                           | The timeout when connecting to MongoDB (`time.Duration` format)                                                  |
| MONGODB_QUERY_TIMEOUT        | 15s                                                          | The timeout for querying MongoDB (`time.Duration` format)                                                        |
| MONGODB_IS_SSL               | false                                                        | Switch to use (or not) TLS when connecting to mongodb                                                            |
| SHUTDOWN_TIMEOUT             | 5s                                                           | The graceful shutdown timeout (`time.Duration` format)                                                           |
| DATASET_API_URL              | http://localhost:22000                                       | The URL of the Dataset API                                                                                       |
| HEALTHCHECK_INTERVAL         | 30s                                                          | Time between self-healthchecks (`time.Duration` format)                                                          |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                                                          | The time taken for the health changes from warning state to critical due to subsystem check failures             |
| OTEL_BATCH_TIMEOUT           | 5s                                                           | Interval between pushes to OT Collector                                                                          |
| OTEL_EXPORTER_OTLP_ENDPOINT  | http://localhost:4317                                        | URL for OpenTelemetry endpoint                                                                                   |
| OTEL_SERVICE_NAME            | "dp-filter-api"                                              | Service name to report to telemetry tools                                                                        |
| SERVICE_AUTH_TOKEN           | FD0108EA-825D-411C-9B1D-41EF7727F465                         | The token used to identify this service when authenticating                                                      |
| ZEBEDEE_URL                  | "http://localhost:8082"                                      | Zebedee URL                                                                                                      |
| ENABLE_PRIVATE_ENDPOINTS     | false                                                        | true if private endpoints should be enabled                                                                      |
| DOWNLOAD_SERVICE_URL         | http://localhost:23600                                       | The URL of the download service                                                                                  |
| DOWNLOAD_SERVICE_SECRET_KEY  | QB0108EZ-825D-412C-9B1D-41EF7747F462                         | The service token for the download service                                                                       |
| ENABLE_URL_REWRITING         | false                                                        | Feature flag to enable URL rewriting                                                                             |
| INTERNAL_DOWNLOAD_SERVICE_URL| "http://localhost:23600"                                     | Internal download service URL                                                                                    |
| EXTERNAL_DOWNLOAD_SERVICE_URL| "http://localhost:23600"                                     | External download service URL                                                                                    |

**Notes:**

1. For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)

### Healthchecking

Currently checked each `$HEALTHCHECK_INTERVAL` and reported on endpoint `/healthcheck`:

* Neo4j
* MongoDB
* Dataset API

### Acceptance tests

To run the acceptance tests for this project, use the following commands

* Start Kafka and Mongodb
* Run the dp-filter-api using `make acceptance`
* Run the tests in dp-api-test

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
