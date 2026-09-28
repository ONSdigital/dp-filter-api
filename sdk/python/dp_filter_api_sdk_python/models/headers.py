from pydantic import BaseModel


class HTTPHeaders(BaseModel):
    access_token: str | None = None
    collection_id: str | None = None
    download_service_token: str | None = None
    florence_token: str | None = None
    if_match: str | None = None

    def to_http_headers(self) -> dict[str, str]:
        headers: dict[str, str] = {}

        if self.access_token is not None:
            headers["Authorization"] = f"Bearer {self.access_token}"
        if self.collection_id is not None:
            headers["Collection-Id"] = self.collection_id
        if self.download_service_token is not None:
            headers["X-Download-Service-Token"] = self.download_service_token
        if self.florence_token is not None:
            headers["X-Florence-Token"] = self.florence_token
        if self.if_match is not None:
            headers["If-Match"] = self.if_match

        return headers
