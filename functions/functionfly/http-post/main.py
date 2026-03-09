import urllib.request
import urllib.error
import json
from urllib.parse import urlencode


def handler(event):
    url = event.get("url")
    data = event.get("data")
    headers = event.get("headers", {})
    content_type = event.get("content_type", "application/json")
    timeout = event.get("timeout", 30)
    follow_redirects = event.get("follow_redirects", True)

    if not url:
        return {"ok": False, "error": "url is required"}

    if not isinstance(url, str):
        return {"ok": False, "error": "url must be a string"}

    if not isinstance(headers, dict):
        return {"ok": False, "error": "headers must be an object"}

    try:
        # Prepare request data
        request_data = None
        if data is not None:
            if content_type == "application/json":
                request_data = json.dumps(data).encode('utf-8')
                headers["Content-Type"] = "application/json"
            elif content_type == "application/x-www-form-urlencoded":
                if isinstance(data, dict):
                    request_data = urlencode(data).encode('utf-8')
                else:
                    request_data = str(data).encode('utf-8')
                headers["Content-Type"] = "application/x-www-form-urlencoded"
            else:
                # Raw data
                if isinstance(data, str):
                    request_data = data.encode('utf-8')
                elif isinstance(data, (bytes, bytearray)):
                    request_data = data
                else:
                    request_data = json.dumps(data).encode('utf-8')
                headers["Content-Type"] = content_type

        # Set default User-Agent if not provided
        if "User-Agent" not in headers:
            headers["User-Agent"] = "FunctionFly/1.0"

        req = urllib.request.Request(url, data=request_data, headers=headers, method="POST")

        with urllib.request.urlopen(req, timeout=timeout) as response:
            # Read response
            content = response.read()
            response_content_type = response.headers.get('Content-Type', '')

            # Try to decode based on content-type
            try:
                if 'application/json' in response_content_type:
                    result = json.loads(content.decode('utf-8'))
                else:
                    result = content.decode('utf-8')
            except (UnicodeDecodeError, json.JSONDecodeError):
                # Return raw bytes as base64 if decoding fails
                import base64
                result = base64.b64encode(content).decode('utf-8')
                response_content_type = "application/octet-stream"

            return {
                "ok": True,
                "status_code": response.status,
                "headers": dict(response.headers),
                "content_type": response_content_type,
                "content_length": len(content),
                "result": result,
                "url": response.url
            }

    except urllib.error.HTTPError as e:
        try:
            error_content = e.read().decode('utf-8')
            error_data = json.loads(error_content) if error_content else None
        except:
            error_data = None

        return {
            "ok": False,
            "error": f"HTTP {e.code}: {e.reason}",
            "status_code": e.code,
            "headers": dict(e.headers) if e.headers else {},
            "error_data": error_data
        }
    except urllib.error.URLError as e:
        return {"ok": False, "error": f"URL Error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Request failed: {str(e)}"}