from dp_filter_api_sdk_python.client import FilterAPIClient
from dp_filter_api_sdk_python.models import (
    CreateFilterRequest,
    CreateFilterResult,
    Dataset,
    Dimension,
    DimensionOptionsResult,
    FilterOutput,
)

from .fakes import FakeResponse, FakeSession


def test_get_filter_output_parses_filter_api_response() -> None:
    payload = {
        "id": "output-123",
        "filter_id": "filter-123",
        "instance_id": "instance-123",
        "dataset": {
            "id": "cpih01",
            "edition": "time-series",
            "version": 2,
        },
        "published": True,
        "state": "completed",
        "dimensions": [{"name": "time", "options": ["2024"]}],
        "downloads": {
            "csv": {
                "href": "http://localhost:23600/downloads/filter-outputs/output-123.csv",
                "public": "public-link",
                "size": "12mb",
            }
        },
        "links": {
            "self": {"href": "http://localhost:22100/filter-outputs/output-123"},
            "filter_blueprint": {"href": "http://localhost:22100/filters/filter-123"},
        },
    }
    session = FakeSession(FakeResponse(200, payload))
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )

    result = client.filter.get_filter_output("output-123")

    assert isinstance(result, FilterOutput)
    assert result.id == "output-123"
    assert result.dataset is not None
    assert result.dataset.id == "cpih01"
    assert result.downloads is not None
    assert result.downloads.csv is not None
    assert (
        result.downloads.csv.href
        == "http://localhost:23600/downloads/filter-outputs/output-123.csv"
    )
    assert session.last_kwargs["method"] == "GET"
    assert session.last_kwargs["url"] == (
        "http://localhost:22100/filter-outputs/output-123"
    )


def test_get_filter_output_parses_filter_flex_api_response() -> None:
    payload = {
        "id": "output-456",
        "filter_id": "filter-456",
        "instance_id": "instance-456",
        "dataset": {
            "id": "population",
            "edition": "2021",
            "version": 1,
            "title": "Population",
            "lowest_geography": "England",
            "release_date": "2025-01-15",
        },
        "published": True,
        "custom": True,
        "state": "completed",
        "type": "cantabular",
        "population_type": "Teaching-Dataset",
        "dimensions": [
            {
                "name": "geography",
                "id": "geography",
                "label": "Geography",
                "options": ["E92000001"],
                "default_categorisation": "area",
            }
        ],
        "downloads": {
            "xlsx": {
                "href": "http://localhost:23600/downloads/output-456.xlsx",
                "size": "24mb",
            }
        },
        "events": [
            {
                "name": "FilterOutputCompleted",
                "timestamp": "2025-01-15T12:00:00Z",
            }
        ],
        "disclosure_control": {
            "status": "completed",
            "dimension": "geography",
            "blocked_options": {
                "blocked_options": ["E92000001"],
                "blocked_count": 1,
            },
        },
        "links": {
            "self": {"href": "http://localhost:22100/filter-outputs/output-456"},
            "filter_blueprint": {"href": "http://localhost:22100/filters/filter-456"},
        },
    }
    session = FakeSession(FakeResponse(200, payload))
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )

    result = client.filter.get_filter_output("output-456")

    assert isinstance(result, FilterOutput)
    assert result.custom is True
    assert result.population_type == "Teaching-Dataset"
    assert result.dataset is not None
    assert result.dataset.title == "Population"
    assert result.dimensions[0].default_categorisation == "area"
    assert result.downloads is not None
    assert result.downloads.xlsx is not None
    assert (
        result.downloads.xlsx.href == "http://localhost:23600/downloads/output-456.xlsx"
    )
    assert result.disclosure_control is not None
    assert result.disclosure_control.blocked_options is not None
    assert result.disclosure_control.blocked_options.blocked_count == 1
    assert result.events[0].name == "FilterOutputCompleted"
    assert session.last_kwargs["method"] == "GET"
    assert session.last_kwargs["url"] == (
        "http://localhost:22100/filter-outputs/output-456"
    )


def test_get_dimension_options_parses_filter_api_response() -> None:
    payload = {
        "items": [
            {
                "option": "2014",
                "links": {
                    "self": {
                        "id": "2014",
                        "href": (
                            "http://localhost:22100/filters/filter-123/dimensions/time/options/2014"
                        ),
                    },
                    "filter": {
                        "id": "filter-123",
                        "href": "http://localhost:22100/filters/filter-123",
                    },
                    "dimension": {
                        "id": "time",
                        "href": (
                            "http://localhost:22100/filters/filter-123/dimensions/time"
                        ),
                    },
                },
            }
        ],
        "count": 1,
        "offset": 1,
        "limit": 1,
        "total_count": 10,
    }
    session = FakeSession(
        FakeResponse(200, payload, headers={"ETag": "etag-filter-api"})
    )
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )

    result = client.filter.get_dimension_options(
        "filter-123",
        "time",
        offset=1,
        limit=1,
    )

    assert isinstance(result, DimensionOptionsResult)
    assert result.etag == "etag-filter-api"
    assert result.options.items[0].option == "2014"
    assert result.options.offset == 1
    assert result.options.limit == 1
    assert result.options.items[0].links is not None
    assert result.options.items[0].links.dimension is not None
    assert result.options.items[0].links.dimension.id == "time"
    assert session.last_kwargs["method"] == "GET"
    assert session.last_kwargs["url"] == (
        "http://localhost:22100/filters/filter-123/dimensions/time/options"
    )
    assert session.last_kwargs["params"] == {"offset": 1, "limit": 1}


def test_get_dimension_options_parses_filter_flex_api_response() -> None:
    payload = {
        "items": [
            {
                "option": "E92000001",
                "links": {
                    "self": {
                        "id": "E92000001",
                        "href": (
                            "http://localhost:22100/filters/filter-456/dimensions/geography/options/E92000001"
                        ),
                    },
                    "filter": {
                        "id": "filter-456",
                        "href": "http://localhost:22100/filters/filter-456",
                    },
                    "Dimension": {
                        "id": "geography",
                        "href": (
                            "http://localhost:22100/filters/filter-456/dimensions/geography"
                        ),
                    },
                },
            }
        ],
        "count": 1,
        "offset": 1,
        "limit": 1,
        "total_count": 25,
    }
    session = FakeSession(
        FakeResponse(200, payload, headers={"ETag": "etag-filter-flex"})
    )
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )

    result = client.filter.get_dimension_options(
        "filter-456",
        "geography",
        offset=1,
        limit=1,
    )

    assert isinstance(result, DimensionOptionsResult)
    assert result.etag == "etag-filter-flex"
    assert result.options.items[0].option == "E92000001"
    assert result.options.items[0].links is not None
    assert result.options.items[0].links.dimension is not None
    assert result.options.items[0].links.dimension.id == "geography"
    assert session.last_kwargs["method"] == "GET"
    assert session.last_kwargs["url"] == (
        "http://localhost:22100/filters/filter-456/dimensions/geography/options"
    )
    assert session.last_kwargs["params"] == {"offset": 1, "limit": 1}


def test_create_filter_posts_request_and_parses_filter_api_response() -> None:
    payload = {
        "filter_id": "filter-123",
        "instance_id": "instance-123",
        "dataset": {
            "id": "cpih01",
            "edition": "time-series",
            "version": 2,
        },
        "dimensions": [{"name": "time"}],
        "state": "created",
        "published": True,
        "links": {"self": {"href": "http://localhost:22100/filters/filter-123"}},
    }
    session = FakeSession(
        FakeResponse(201, payload, headers={"ETag": "etag-filter-api"})
    )
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )
    request = CreateFilterRequest(
        dataset=Dataset(id="cpih01", edition="time-series", version=2),
        dimensions=[Dimension(name="time")],
    )

    result = client.filter.create_filter(request)

    assert isinstance(result, CreateFilterResult)
    assert result.filter.filter_id == "filter-123"
    assert result.filter.dataset is not None
    assert result.filter.dataset.id == "cpih01"
    assert result.etag == "etag-filter-api"

    assert session.last_kwargs["method"] == "POST"
    assert session.last_kwargs["url"] == "http://localhost:22100/filters"
    assert session.last_kwargs["json"] == {
        "dataset": {
            "id": "cpih01",
            "edition": "time-series",
            "version": 2,
        },
        "dimensions": [{"name": "time", "options": []}],
    }


def test_create_filter_posts_flex_fields_and_parses_flex_response() -> None:
    payload = {
        "filter_id": "filter-456",
        "instance_id": "instance-456",
        "dataset": {
            "id": "population",
            "edition": "2021",
            "version": 1,
            "title": "Population",
            "lowest_geography": "England",
            "release_date": "2025-01-15",
        },
        "dimensions": [
            {
                "name": "geography",
                "id": "geography",
                "label": "Geography",
                "options": ["E92000001"],
                "default_categorisation": "area",
            }
        ],
        "state": "created",
        "published": True,
        "custom": True,
        "population_type": "Teaching-Dataset",
        "type": "cantabular",
        "links": {
            "self": {"href": "http://localhost:22100/filters/filter-456"},
            "version": {
                "href": "http://localhost:22000/datasets/population/editions/2021/versions/1"
            },
        },
    }
    session = FakeSession(
        FakeResponse(201, payload, headers={"ETag": "etag-filter-flex"})
    )
    client = FilterAPIClient(
        base_url="http://localhost:22100",
        session=session,
    )
    request = CreateFilterRequest(
        dataset=Dataset(id="population", edition="2021", version=1),
        dimensions=[
            Dimension(name="geography", options=["E92000001"]),
        ],
        population_type="Teaching-Dataset",
        custom=True,
    )

    result = client.filter.create_filter(request)

    assert isinstance(result, CreateFilterResult)
    assert result.filter.filter_id == "filter-456"
    assert result.filter.custom is True
    assert result.filter.population_type == "Teaching-Dataset"
    assert result.filter.dataset is not None
    assert result.filter.dataset.title == "Population"
    assert result.filter.dimensions[0].default_categorisation == "area"
    assert result.etag == "etag-filter-flex"

    assert session.last_kwargs["method"] == "POST"
    assert session.last_kwargs["url"] == "http://localhost:22100/filters"
    assert session.last_kwargs["json"] == {
        "dataset": {
            "id": "population",
            "edition": "2021",
            "version": 1,
        },
        "dimensions": [
            {
                "name": "geography",
                "options": ["E92000001"],
            }
        ],
        "population_type": "Teaching-Dataset",
        "custom": True,
    }
