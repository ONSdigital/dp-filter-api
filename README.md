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

| Environment variable | Default                                   | Description
| -------------------- | ----------------------------------------- | -----------
| BIND_ADDR            | :22100                                    | The host and port to bind to
| POSTGRES_URL         | user=dp dbname=FilterJobs sslmode=disable | URL to a Postgres services
| HOST                 |  "http://localhost:22100"                 | The host name used to build URLs

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
