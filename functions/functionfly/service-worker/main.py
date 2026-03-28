import urllib.request
import urllib.error
import ssl
from urllib.parse import urljoin


def check_url(url, timeout, ctx):
    try:
        req = urllib.request.Request(url, method="HEAD")
        with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
            return resp.status < 400
    except urllib.error.HTTPError as e:
        return e.code < 400
    except Exception:
        return False


def handler(event):
    """Check if a URL has a service worker."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        # Check common service worker paths
        base_url = url.rstrip("/")
        sw_paths = ["/sw.js", "/service-worker.js", "/serviceworker.js", "/sw-prod.js"]

        sw_url = None
        for path in sw_paths:
            candidate = urljoin(base_url + "/", path.lstrip("/"))
            if check_url(candidate, timeout, ctx):
                sw_url = candidate
                break

        return {
            "ok": True,
            "has_service_worker": sw_url is not None,
            "sw_url": sw_url,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
