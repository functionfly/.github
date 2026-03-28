import urllib.request
import urllib.error
import ssl
import re


def parse_max_age(cache_control):
    if not cache_control:
        return None
    match = re.search(r'max-age=(\d+)', cache_control)
    return int(match.group(1)) if match else None


def handler(event):
    """Analyze HTTP caching headers."""
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
                headers = {k.lower(): v for k, v in resp.headers.items()}
        except urllib.error.HTTPError as e:
            headers = {k.lower(): v for k, v in e.headers.items()}

        cache_control = headers.get("cache-control")
        etag = headers.get("etag")
        last_modified = headers.get("last-modified")
        expires = headers.get("expires")
        pragma = headers.get("pragma")
        vary = headers.get("vary")

        max_age = parse_max_age(cache_control)
        no_cache = cache_control and ("no-cache" in cache_control or "no-store" in cache_control)
        cacheable = not no_cache and (cache_control is not None or etag is not None or last_modified is not None)

        return {
            "ok": True,
            "cacheable": cacheable,
            "cache_control": cache_control,
            "etag": etag,
            "last_modified": last_modified,
            "expires": expires,
            "pragma": pragma,
            "vary": vary,
            "max_age": max_age,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
