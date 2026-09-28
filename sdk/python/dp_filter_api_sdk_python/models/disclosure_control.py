from pydantic import BaseModel, Field


class BlockedOptions(BaseModel):
    blocked_options: list[str] = Field(default_factory=list)
    blocked_count: int = 0


class DisclosureControl(BaseModel):
    status: str | None = None
    dimension: str | None = None
    blocked_options: BlockedOptions | None = None
