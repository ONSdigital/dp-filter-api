import pytest
from pydantic import ValidationError

from dp_filter_api_sdk_python.models import PaginationParams


def test_pagination_params_converts_set_values_to_query_params() -> None:
    params = PaginationParams(offset=1, limit=1)

    assert params.to_query_params() == {"offset": 1, "limit": 1}


def test_pagination_params_omits_unset_values() -> None:
    params = PaginationParams()

    assert params.to_query_params() == {}


@pytest.mark.parametrize(
    ("field", "value"),
    [
        ("offset", -1),
        ("limit", -1),
        ("offset", "1"),
        ("limit", "1"),
    ],
)
def test_pagination_params_rejects_invalid_values(field: str, value: int | str) -> None:
    with pytest.raises(ValidationError):
        PaginationParams(**{field: value})
