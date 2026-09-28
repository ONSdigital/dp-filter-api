from dp_filter_api_sdk_python.models import HTTPHeaders


def test_to_http_headers_converts_all_values() -> None:
    headers = HTTPHeaders(
        access_token="access-token",
        collection_id="collection-id",
        download_service_token="download-service-token",
        florence_token="florence-token",
        if_match="etag-value",
    )

    assert headers.to_http_headers() == {
        "Authorization": "Bearer access-token",
        "Collection-Id": "collection-id",
        "X-Download-Service-Token": "download-service-token",
        "X-Florence-Token": "florence-token",
        "If-Match": "etag-value",
    }


def test_to_http_headers_omits_unset_values() -> None:
    headers = HTTPHeaders(access_token="access-token")

    assert headers.to_http_headers() == {
        "Authorization": "Bearer access-token",
    }
