import urllib.request
import urllib.error
import ssl


def handler(event):
    """Check the size of a JavaScript bundle."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        req = urllib.request.Request(url, method="HEAD")
        req.add_header("Accept-Encoding", "gzip, br")
        try:
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = {k.lower(): v for k, v in resp.headers.items()}
        except urllib.error.HTTPError as e:
            headers = {k.lower(): v for k, v in e.headers.items()}

        content_length = headers.get("content-length")
        size_bytes = int(content_length) if content_length else None
        size_kb = round(size_bytes / 1024, 2) if size_bytes else None
        encoding = headers.get("content-encoding", "")

        suggestions = []
        if size_bytes:
            if size_bytes > 250 * 1024:
                suggestions.append(f"Bundle is {size_kb}KB - consider code splitting (target: <250KB)")
            if size_bytes > 100 * 1024:
                suggestions.append("Consider tree shaking to remove unused code")
            if not encoding:
                suggestions.append("Enable gzip or Brotli compression on the server")
            if size_bytes > 50 * 1024:
                suggestions.append("Consider lazy loading non-critical JavaScript")

        return {
            "ok": True,
            "url": url,
            "size_bytes": size_bytes,
            "size_kb": size_kb,
            "encoding": encoding or None,
            "suggestions": suggestions,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
