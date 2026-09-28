from datetime import datetime

from pydantic import BaseModel


class Dataset(BaseModel):
    id: str
    edition: str
    version: int

    # These fields are returned when proxied to dp-cantabular-filter-flex-api.
    lowest_geography: str | None = None
    release_date: str | None = None
    title: str | None = None


class LinkObject(BaseModel):
    id: str | None = None
    href: str | None = None

    # This field is returned when proxied to dp-cantabular-filter-flex-api.
    label: str | None = None


class DownloadItem(BaseModel):
    skipped: bool = False
    href: str | None = None
    private: str | None = None
    public: str | None = None
    size: str | None = None


class Downloads(BaseModel):
    csv: DownloadItem | None = None
    xls: DownloadItem | None = None

    # These formats are returned when proxied to dp-cantabular-filter-flex-api.
    xlsx: DownloadItem | None = None
    txt: DownloadItem | None = None
    csvw: DownloadItem | None = None


class Event(BaseModel):
    type: str | None = None
    time: datetime | None = None

    # These fields are returned when proxied to dp-cantabular-filter-flex-api.
    name: str | None = None
    timestamp: str | None = None
