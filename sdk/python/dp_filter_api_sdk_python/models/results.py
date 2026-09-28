from pydantic import BaseModel

from .filter import Filter


class CreateFilterResult(BaseModel):
    filter: Filter
    etag: str | None = None
