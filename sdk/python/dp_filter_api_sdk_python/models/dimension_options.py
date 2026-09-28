from pydantic import AliasChoices, BaseModel, Field

from .common import LinkObject


class DimensionOptionLinks(BaseModel):
    self: LinkObject | None = None
    filter: LinkObject | None = None

    # dp-filter-api returns "dimension" while dp-cantabular-filter-flex-api may return "Dimension".
    dimension: LinkObject | None = Field(
        default=None,
        validation_alias=AliasChoices("dimension", "Dimension"),
    )


class DimensionOption(BaseModel):
    links: DimensionOptionLinks | None = None
    option: str


class DimensionOptions(BaseModel):
    items: list[DimensionOption] = Field(default_factory=list)
    count: int = 0
    offset: int = 0
    limit: int = 0
    total_count: int = 0


class DimensionOptionsResult(BaseModel):
    options: DimensionOptions
    etag: str | None = None
