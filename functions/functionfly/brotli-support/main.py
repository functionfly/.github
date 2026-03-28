import urllib.request
import urllib.error
import ssl


def handler(event):
    """Check if a server supports Brotli content encoding."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        req = urllib.request.Request(url)
        req.add_header("Accept-Encoding", "br, gzip, deflate")

        try:
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = {k.lower(): v for k, v in resp.headers.items()}
        except urllib.error.HTTPError as e:
            headers = {k.lower(): v for k, v in e.headers.items()}

        encoding = headers.get("content-encoding", "")
        supported = "br" in encoding.lower()

        return {
            "ok": True,
            "supported": supported,
            "encoding": encoding or None,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
