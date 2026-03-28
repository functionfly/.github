import urllib.request
import urllib.error
import ssl


def handler(event):
    """Check if a server advertises HTTP/3 support via Alt-Svc header."""
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
                headers = dict(resp.headers)
        except urllib.error.HTTPError as e:
            headers = dict(e.headers)

        alt_svc = headers.get("alt-svc") or headers.get("Alt-Svc")
        supported = alt_svc is not None and ("h3" in alt_svc or "quic" in alt_svc.lower())

        return {
            "ok": True,
            "supported": supported,
            "alt_svc": alt_svc,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
