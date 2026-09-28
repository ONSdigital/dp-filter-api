from pydantic import BaseModel, Field


class PaginationParams(BaseModel):
    offset: int | None = Field(default=None, ge=0, strict=True)
    limit: int | None = Field(default=None, ge=0, strict=True)

    def to_query_params(self) -> dict[str, int]:
        return self.model_dump(exclude_none=True)
