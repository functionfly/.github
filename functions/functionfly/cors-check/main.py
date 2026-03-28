import urllib.request
import urllib.error
import ssl


def handler(event):
    """Check CORS configuration for a URL."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        origin = event.get("origin", "https://example.com")
        method = event.get("method", "GET")
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        # Send OPTIONS preflight request
        req = urllib.request.Request(url, method="OPTIONS")
        req.add_header("Origin", origin)
        req.add_header("Access-Control-Request-Method", method)
        req.add_header("Access-Control-Request-Headers", "Content-Type")

        try:
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = dict(resp.headers)
        except urllib.error.HTTPError as e:
            headers = dict(e.headers)

        allow_origin = headers.get("access-control-allow-origin") or headers.get("Access-Control-Allow-Origin")
        allow_methods = headers.get("access-control-allow-methods") or headers.get("Access-Control-Allow-Methods")
        allow_headers = headers.get("access-control-allow-headers") or headers.get("Access-Control-Allow-Headers")
        allow_credentials = headers.get("access-control-allow-credentials") or headers.get("Access-Control-Allow-Credentials")

        cors_enabled = allow_origin is not None

        return {
            "ok": True,
            "cors_enabled": cors_enabled,
            "allow_origin": allow_origin,
            "allow_methods": allow_methods,
            "allow_headers": allow_headers,
            "allow_credentials": allow_credentials,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
