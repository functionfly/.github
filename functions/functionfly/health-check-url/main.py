import urllib.request
import urllib.error
import time
import ssl


def handler(event):
    """Perform an HTTP health check on a URL."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))
        expected_status = int(event.get("expected_status", 200))
        method = event.get("method", "GET").upper()

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        start = time.time()
        try:
            req = urllib.request.Request(url, method=method)
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                status_code = resp.status
                latency_ms = round((time.time() - start) * 1000, 2)
                healthy = status_code == expected_status
                return {
                    "ok": True,
                    "healthy": healthy,
                    "status_code": status_code,
                    "latency_ms": latency_ms,
                    "url": url,
                }
        except urllib.error.HTTPError as e:
            latency_ms = round((time.time() - start) * 1000, 2)
            healthy = e.code == expected_status
            return {
                "ok": True,
                "healthy": healthy,
                "status_code": e.code,
                "latency_ms": latency_ms,
                "url": url,
            }
        except urllib.error.URLError as e:
            latency_ms = round((time.time() - start) * 1000, 2)
            return {
                "ok": True,
                "healthy": False,
                "status_code": None,
                "latency_ms": latency_ms,
                "url": url,
                "error": str(e.reason),
            }
    except Exception as e:
        return {"ok": False, "error": str(e)}
