from collections.abc import Mapping
from typing import Any, Protocol, runtime_checkable

import requests

from .models import (
    CreateFilterRequest,
    CreateFilterResult,
    DimensionOptionsResult,
    FilterOutput,
)


@runtime_checkable
class Headers(Protocol):
    def to_http_headers(self) -> Mapping[str, str]: ...


class RequestingClient(Protocol):
    def _request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json: dict[str, Any] | None = None,
        headers: Mapping[str, str] | None = None,
    ) -> requests.Response: ...


@runtime_checkable
class FilterClientProtocol(Protocol):
    """Protocol for filter-specific operations."""

    def get_filter_output(
        self,
        filter_output_id: str,
        headers: Headers | None = None,
    ) -> FilterOutput: ...

    def get_dimension_options(
        self,
        filter_id: str,
        dimension_name: str,
        headers: Headers | None = None,
        offset: int | None = None,
        limit: int | None = None,
    ) -> DimensionOptionsResult: ...

    def create_filter(
        self,
        create_filter_request: CreateFilterRequest,
        headers: Headers | None = None,
    ) -> CreateFilterResult: ...


@runtime_checkable
class FilterAPIClientProtocol(Protocol):
    """Protocol for the FilterAPIClient."""

    def health(self) -> dict[str, Any]: ...

    filter: FilterClientProtocol
