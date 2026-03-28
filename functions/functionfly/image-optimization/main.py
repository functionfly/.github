import urllib.request
import urllib.error
import ssl


def handler(event):
    """Analyze an image URL for optimization opportunities."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        req = urllib.request.Request(url, method="HEAD")
        try:
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = {k.lower(): v for k, v in resp.headers.items()}
        except urllib.error.HTTPError as e:
            headers = {k.lower(): v for k, v in e.headers.items()}

        content_type = headers.get("content-type", "")
        content_length = headers.get("content-length")
        size_bytes = int(content_length) if content_length else None

        # Determine format
        fmt = None
        if "jpeg" in content_type or "jpg" in content_type:
            fmt = "jpeg"
        elif "png" in content_type:
            fmt = "png"
        elif "gif" in content_type:
            fmt = "gif"
        elif "webp" in content_type:
            fmt = "webp"
        elif "svg" in content_type:
            fmt = "svg"
        elif "avif" in content_type:
            fmt = "avif"
        else:
            # Try from URL
            url_lower = url.lower()
            for ext in ["webp", "avif", "jpeg", "jpg", "png", "gif", "svg"]:
                if f".{ext}" in url_lower:
                    fmt = ext
                    break

        suggestions = []
        if fmt in ("png", "jpeg", "jpg"):
            suggestions.append("Consider converting to WebP for 25-35% better compression")
        if fmt == "png" and size_bytes and size_bytes > 50000:
            suggestions.append("Large PNG detected - consider lossless compression with pngquant")
        if fmt in ("jpeg", "jpg") and size_bytes and size_bytes > 100000:
            suggestions.append("Large JPEG - consider reducing quality to 80-85%")
        if fmt == "gif":
            suggestions.append("Consider converting animated GIFs to WebM/MP4 video")
        if not headers.get("cache-control"):
            suggestions.append("Add Cache-Control header for browser caching")
        if not (headers.get("content-encoding") or "").lower() in ("gzip", "br"):
            if fmt == "svg":
                suggestions.append("SVG files can be gzip compressed")

        return {
            "ok": True,
            "url": url,
            "format": fmt,
            "size_bytes": size_bytes,
            "content_type": content_type,
            "suggestions": suggestions,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
