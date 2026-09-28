# dp-filter-api-sdk-python

A Python SDK client for `dp-filter-api`.

## Requirements

- Python `>=3.14,<3.15`
- Poetry `>=2.0.0,<3.0.0` (for local development)

## Install

### Install from a release tag

```bash
pip install "git+https://github.com/ONSdigital/dp-filter-api.git@<release-tag>#subdirectory=sdk/python"
```

Release tags can be found in the [GitHub releases](https://github.com/ONSdigital/dp-filter-api/releases) page.

### Install for local development

```bash
make install-dev
```

## Quick Start

Create a client and get a filter output.

```python
from dp_filter_api_sdk_python import HTTPHeaders, create_client

client = create_client(
    base_url="http://localhost:22100",
)

headers = HTTPHeaders(
    access_token="<service-auth-token>",
)

output = client.filter.get_filter_output(
    filter_output_id="<filter-output-id>",
    headers=headers,
)

print(output.filter_id)
print(output.state)
```

## Create a Client

Use `create_client()` to create an instance of the SDK client. By default, the client creates a `requests.Session` and uses a 10 second timeout.

### Let the SDK create a session

```python
from dp_filter_api_sdk_python import create_client

client = create_client(
    base_url="http://localhost:22100",
)
```

### Pass an existing session

```python
import requests

from dp_filter_api_sdk_python import create_client

session = requests.Session()
client = create_client(
    base_url="http://localhost:22100",
    session=session,
)
```

## Authentication and Headers

Pass an instance of `HTTPHeaders` to the client methods when a request requires authentication or additional headers. The supported headers include:

| `HTTPHeaders` field      | HTTP header                |
| ------------------------ | -------------------------- |
| `florence_token`         | `X-Florence-Token`         |
| `access_token`           | `Authorization`            |
| `collection_id`          | `Collection-Id`            |
| `download_service_token` | `X-Download-Service-Token` |
| `if_match`               | `If-Match`                 |

> **Note:** The `access_token` should be set without the `"Bearer "` prefix.

## Using the Filter Client

## Overview

| SDK method                                                        | HTTP endpoint                                                  | Can be forwarded to dp-cantabular-filter-flex-api? |
| ----------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------- |
| [`client.filter.get_filter_output()`](#get-a-filter-output)       | `GET /filter-outputs/{filter_output_id}`                       | Yes, for flexible or multivariate filters          |
| [`client.filter.get_dimension_options()`](#get-dimension-options) | `GET /filters/{filter_id}/dimensions/{dimension_name}/options` | Yes, for flexible or multivariate filters          |
| [`client.filter.create_filter()`](#create-a-filter)               | `POST /filters`                                                | Yes, for flexible or multivariate datasets         |
| [`client.health()`](#health-check)                                | `GET /health`                                                  | No                                                 |

> **Note:** All SDK requests are sent to dp-filter-api. Depending on the dataset or filter type and the service configuration, dp-filter-api may forward requests for Cantabular flexible or multivariate filters to [dp-cantabular-filter-flex-api](https://github.com/ONSdigital/dp-cantabular-filter-flex-api).

### Get a filter output

```python
output = client.filter.get_filter_output(
    filter_output_id="output-123",
)
```

### Get dimension options

```python
result = client.filter.get_dimension_options(
    filter_id="filter-123",
    dimension_name="time",
)
```

### Create a filter

```python
from dp_filter_api_sdk_python import CreateFilterRequest, Dataset, Dimension

request = CreateFilterRequest(
    dataset=Dataset(
        id="cpih01",
        edition="time-series",
        version=2,
    ),
    dimensions=[Dimension(name="time")],
)

result = client.filter.create_filter(request)
```

To create a filter for a flexible dataset, the request must include the `population_type` and `custom` fields:

```python
from dp_filter_api_sdk_python import CreateFilterRequest, Dataset, Dimension

request = CreateFilterRequest(
    dataset=Dataset(
        id="population",
        edition="2021",
        version=1,
    ),
    dimensions=[
        Dimension(
            name="geography",
            id="geography",
            options=["E92000001"],
            is_area_type=True,
        ),
    ],
    population_type="Teaching-Dataset",
    custom=True,
)

result = client.filter.create_filter(request)
```

### Health check

```python
health = client.health()
print(health)
```

## Error Handling

The SDK raises typed exceptions for common HTTP errors. They inherit from `APIError`, which exposes the HTTP status code as `status_code`.

```python
from dp_filter_api_sdk_python import (
    APIError,
    AuthenticationError,
    ConflictError,
    NotFoundError,
    UnprocessableEntityError,
    ValidationError,
    create_client,
)

client = create_client(
    base_url="http://localhost:22100",
)

try:
    output = client.filter.get_filter_output("output-123")
except AuthenticationError as error:
    print(f"Authentication failed: {error}; status={error.status_code}")
except NotFoundError as error:
    print(f"Resource not found: {error}")
except ValidationError as error:
    print(f"Request validation failed: {error}")
except ConflictError as error:
    print(f"Request conflicts with the resource: {error}")
except UnprocessableEntityError as error:
    print(f"Request could not be processed: {error}")
except APIError as error:
    print(f"API request failed: {error}; status={error.status_code}")
```

> **Note:**: Network errors are also wrapped as an `APIError` but the `status_code` is `None`.

## Development

See the [Makefile](Makefile) for available development commands.
