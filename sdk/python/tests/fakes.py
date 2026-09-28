import requests


class FakeResponse:
    def __init__(
        self,
        status_code: int,
        payload: dict | None = None,
        headers: dict[str, str] | None = None,
    ) -> None:
        self.status_code = status_code
        self._payload = payload if payload is not None else {}
        self.content = b"{}" if payload is not None else b""
        self.text = str(self._payload)
        self.headers: dict[str, str] = dict(headers) if headers is not None else {}

    def json(self) -> dict:
        return self._payload


class FakeSession(requests.Session):
    def __init__(
        self,
        response: FakeResponse | None = None,
        error: requests.RequestException | None = None,
    ) -> None:
        super().__init__()
        self.response = response
        self.error = error
        self.last_kwargs: dict = {}

    def request(self, **kwargs):
        self.last_kwargs = kwargs

        if self.error is not None:
            raise self.error

        if self.response is None:
            raise AssertionError("FakeSession has no response set")

        return self.response
