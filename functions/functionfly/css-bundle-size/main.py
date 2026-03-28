import urllib.request
import urllib.error
import ssl


def handler(event):
    """Check the size of a CSS bundle."""
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
            if size_bytes > 50 * 1024:
                suggestions.append(f"CSS is {size_kb}KB - consider critical CSS extraction")
            if size_bytes > 20 * 1024:
                suggestions.append("Use PurgeCSS to remove unused styles")
            if not encoding:
                suggestions.append("Enable gzip or Brotli compression")
            suggestions.append("Consider CSS minification if not already applied")

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
