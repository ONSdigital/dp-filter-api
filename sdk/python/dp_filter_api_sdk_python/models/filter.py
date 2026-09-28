from __future__ import annotations

from pydantic import BaseModel, Field

from .common import Dataset, Downloads, Event, LinkObject
from .disclosure_control import DisclosureControl


class Dimension(BaseModel):
    name: str
    id: str | None = None
    label: str | None = None
    dimension_url: str | None = None

    # These fields are returned when proxied to dp-cantabular-filter-flex-api.
    default_categorisation: str | None = None
    filter_by_parent: str | None = None
    quality_statement_text: str | None = None
    quality_summary_url: str | None = None
    options: list[str] = Field(default_factory=list)
    is_area_type: bool | None = None


class LinkMap(BaseModel):
    dimensions: LinkObject | None = None
    filter_output: LinkObject | None = None
    filter_blueprint: LinkObject | None = None
    self: LinkObject | None = None
    version: LinkObject | None = None


class Filter(BaseModel):
    id: str | None = None
    dataset: Dataset | None = None
    instance_id: str | None = None
    dimensions: list[Dimension] = Field(default_factory=list)
    downloads: Downloads | None = None
    events: list[Event] = Field(default_factory=list)
    filter_id: str | None = None
    state: str | None = None
    published: bool | None = None

    # These fields are returned when proxied to dp-cantabular-filter-flex-api.
    custom: bool | None = None
    population_type: str | None = None
    disclosure_control: DisclosureControl | None = None
    links: LinkMap | None = None
    type: str | None = None


class FilterOutput(BaseModel):
    id: str | None = None
    filter_id: str | None = None
    instance_id: str | None = None
    dataset: Dataset | None = None
    published: bool | None = None
    state: str | None = None
    downloads: Downloads | None = None
    events: list[Event] = Field(default_factory=list)
    type: str | None = None

    # These fields are returned when proxied to dp-cantabular-filter-flex-api.
    custom: bool | None = None
    population_type: str | None = None
    disclosure_control: DisclosureControl | None = None
    links: LinkMap | None = None
    dimensions: list[Dimension] = Field(default_factory=list)
