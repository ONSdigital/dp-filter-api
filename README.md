dp-filter-api
================

### Getting started

#### mongo
* Run ```brew install mongodb```
* Run ```brew services start mongodb```

#### Kafka
* Run ```brew install zookeeper ```
* Download kafka version 0.10.2.1
* Run ```./kafka-server-start.sh ../config/server.properties```

### Configuration

| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | :22100                                    | The host and port to bind to
| HOST                       | http://localhost:22100                    | The host name used to build URLs
| KAFKA_ADDR                 | localhost:9092                            | The kafka broker addresses (can be comma separated)
| FILTER_JOB_SUBMITTED_TOPIC | filter-job-submitted                      | The kafka topic to write messages to
| KAFKA_MAX_BYTES            | 2000000                                   | The maximum permitted size of a message. Should be set equal to or smaller than the broker's `message.max.bytes`
| MONGODB_BIND_ADDR          | localhost:27017                           | URL to a mongodb services
| SECRET_KEY                 | FD0108EA-825D-411C-9B1D-41EF7727F465      | A secret key used authentication
| SHUTDOWN_TIMEOUT           | 5                                         | The graceful shutdown timeout in seconds
| DATASET_API_URL            | http://localhost:22000                    | The URL of the Dataset API
| DATASET_API_AUTH_TOKEN     | FD0108EA-825D-411C-9B1D-41EF7727F465      | The token used to access the Dataset API

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
