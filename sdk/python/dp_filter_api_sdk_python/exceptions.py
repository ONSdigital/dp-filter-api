"""Custom exception types used by the SDK."""


class APIError(Exception):
    """Base API error raised for non-successful API responses."""

    def __init__(self, message: str, status_code: int | None = None) -> None:
        super().__init__(message)
        self.status_code = status_code


class ValidationError(APIError):
    """Raised when request validation fails."""


class AuthenticationError(APIError):
    """Raised when API authentication or authorization fails."""


class NotFoundError(APIError):
    """Raised when a requested resource does not exist."""


class ConflictError(APIError):
    """Raised when the request conflicts with the current resource state."""


class UnprocessableEntityError(APIError):
    """Raised when the server cannot process an otherwise valid request."""
