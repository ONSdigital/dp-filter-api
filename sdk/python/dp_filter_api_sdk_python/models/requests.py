from pydantic import BaseModel, Field

from .common import Dataset
from .filter import Dimension


class CreateFilterRequest(BaseModel):
    """Request body for POST /filters."""

    dataset: Dataset
    dimensions: list[Dimension] = Field(default_factory=list)

    # Required when dp-filter-api proxies this request to dp-cantabular-filter-flex-api.
    population_type: str | None = None

    # Enables custom-filter behaviour when the request is proxied to dp-cantabular-filter-flex-api.
    custom: bool | None = None
