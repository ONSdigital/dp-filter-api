from .models import (
    CreateFilterRequest,
    CreateFilterResult,
    DimensionOptions,
    DimensionOptionsResult,
    Filter,
    FilterOutput,
    PaginationParams,
)
from .protocols import Headers, RequestingClient


class FilterAPI:
    def __init__(self, client: RequestingClient) -> None:
        self._client = client

    def get_filter_output(
        self,
        filter_output_id: str,
        headers: Headers | None = None,
    ) -> FilterOutput:
        """GET /filter-outputs/{filter_output_id}

        Requests for flexible or multivariate filter types are forwarded to
        dp-cantabular-filter-flex-api when proxying is enabled.
        """

        response = self._client._request(
            "GET",
            f"/filter-outputs/{filter_output_id}",
            headers=headers.to_http_headers() if headers else None,
        )

        return FilterOutput.model_validate(response.json())

    def get_dimension_options(
        self,
        filter_id: str,
        dimension_name: str,
        headers: Headers | None = None,
        offset: int | None = None,
        limit: int | None = None,
    ) -> DimensionOptionsResult:
        """GET /filters/{filter_id}/dimensions/{dimension_name}/options

        Requests for flexible or multivariate filter types are forwarded to
        dp-cantabular-filter-flex-api when proxying is enabled.
        """

        paginationParams = PaginationParams(offset=offset, limit=limit)

        response = self._client._request(
            "GET",
            f"/filters/{filter_id}/dimensions/{dimension_name}/options",
            params=paginationParams.to_query_params(),
            headers=headers.to_http_headers() if headers else None,
        )

        return DimensionOptionsResult(
            options=DimensionOptions.model_validate(response.json()),
            etag=response.headers.get("ETag"),
        )

    def create_filter(
        self,
        create_filter_request: CreateFilterRequest,
        headers: Headers | None = None,
    ) -> CreateFilterResult:
        """
        POST /filters

        Requests for flexible or multivariate Cantabular datasets are forwarded to
        dp-cantabular-filter-flex-api when proxying is enabled.

        For these requests, population_type is required. Set custom=True to create a
        custom flexible filter.
        """

        response = self._client._request(
            "POST",
            "/filters",
            json=create_filter_request.model_dump(exclude_none=True),
            headers=headers.to_http_headers() if headers else None,
        )

        return CreateFilterResult(
            filter=Filter.model_validate(response.json()),
            etag=response.headers.get("ETag"),
        )
