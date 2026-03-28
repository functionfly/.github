import urllib.request
import urllib.error
import ssl


def handler(event):
    """Fetch HTTP response headers from a URL."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        method = event.get("method", "HEAD").upper()
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        req = urllib.request.Request(url, method=method)
        try:
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = dict(resp.headers)
                status_code = resp.status
        except urllib.error.HTTPError as e:
            headers = dict(e.headers)
            status_code = e.code

        return {"ok": True, "status_code": status_code, "headers": headers, "url": url}
    except Exception as e:
        return {"ok": False, "error": str(e)}
