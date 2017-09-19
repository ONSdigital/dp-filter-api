dp-filter-api
================

### Getting started

#### Postgres
* Run ```brew install postgres```
* Run ```brew services start postgres```
* Run ```createuser dp -d -w```
* Run ```createdb --owner dp FilterJobs```
* Run ```psql -U dp FilterJobs -f scripts/InitDatabase.sql```

### Configuration

| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | :22100                                    | The host and port to bind to
| HOST                       | http://localhost:22100                    | The host name used to build URLs
| KAFKA_ADDR                 | localhost:9092                            | The kafka broker addresses (can be comma separated)
| FILTER_JOB_SUBMITTED_TOPIC | filter-job-submitted                      | The kafka topic to write messages to
| KAFKA_MAX_BYTES            | 2000000                | The maximum permitted size of a message. Should be set equal to or smaller than the broker's `message.max.bytes`
| POSTGRES_URL               | user=dp dbname=FilterJobs sslmode=disable | URL to a Postgres services
| SECRET_KEY                 | FD0108EA-825D-411C-9B1D-41EF7727F465      | A secret key used authentication
| SHUTDOWN_TIMEOUT           | 5s                                  | The graceful shutdown timeout in seconds

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
