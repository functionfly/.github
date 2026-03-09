import urllib.request
import urllib.error
import json


def handler(event):
    url = event.get("url")
    headers = event.get("headers", {})
    params = event.get("params")
    timeout = event.get("timeout", 30)
    follow_redirects = event.get("follow_redirects", True)

    if not url:
        return {"ok": False, "error": "url is required"}

    if not isinstance(url, str):
        return {"ok": False, "error": "url must be a string"}

    if not isinstance(headers, dict):
        return {"ok": False, "error": "headers must be an object"}

    if params and isinstance(params, dict):
        # Add query parameters to URL
        from urllib.parse import urlencode, urlparse, urlunparse
        parsed = urlparse(url)
        if parsed.query:
            existing_params = dict(param.split('=', 1) for param in parsed.query.split('&') if '=' in param)
            existing_params.update(params)
            params = existing_params
        query = urlencode(params)
        url = urlunparse((parsed.scheme, parsed.netloc, parsed.path, parsed.params, query, parsed.fragment))

    try:
        # Set default User-Agent if not provided
        if "User-Agent" not in headers:
            headers["User-Agent"] = "FunctionFly/1.0"

        req = urllib.request.Request(url, headers=headers, method="HEAD")

        with urllib.request.urlopen(req, timeout=timeout) as response:
            # HEAD requests don't return a body, only headers
            return {
                "ok": True,
                "status_code": response.status,
                "headers": dict(response.headers),
                "content_type": response.headers.get('Content-Type', ''),
                "content_length": response.headers.get('Content-Length'),
                "last_modified": response.headers.get('Last-Modified'),
                "etag": response.headers.get('ETag'),
                "url": response.url
            }

    except urllib.error.HTTPError as e:
        return {
            "ok": False,
            "error": f"HTTP {e.code}: {e.reason}",
            "status_code": e.code,
            "headers": dict(e.headers) if e.headers else {}
        }
    except urllib.error.URLError as e:
        return {"ok": False, "error": f"URL Error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Request failed: {str(e)}"}