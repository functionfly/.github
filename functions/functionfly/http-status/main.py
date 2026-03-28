import urllib.request
import urllib.error
import ssl
import http.client


STATUS_TEXTS = {
    200: "OK", 201: "Created", 204: "No Content", 301: "Moved Permanently",
    302: "Found", 304: "Not Modified", 400: "Bad Request", 401: "Unauthorized",
    403: "Forbidden", 404: "Not Found", 405: "Method Not Allowed",
    429: "Too Many Requests", 500: "Internal Server Error",
    502: "Bad Gateway", 503: "Service Unavailable", 504: "Gateway Timeout",
}


def handler(event):
    """Get the HTTP status code for a URL."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        try:
            req = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                status_code = resp.status
        except urllib.error.HTTPError as e:
            status_code = e.code

        status_text = STATUS_TEXTS.get(status_code, http.client.responses.get(status_code, "Unknown"))
        success = 200 <= status_code < 300

        return {
            "ok": True,
            "status_code": status_code,
            "status_text": status_text,
            "success": success,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
