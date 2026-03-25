from urllib.parse import urlparse


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    require_tld = event.get("require_tld", True)

    if value is None:
        return {"ok": False, "error": "value is required"}

    try:
        parsed = urlparse(str(value))
        has_scheme = parsed.scheme in ("http", "https", "ftp", "ftps")
        has_netloc = bool(parsed.netloc)
        has_tld = "." in parsed.netloc if require_tld else True
        result = has_scheme and has_netloc and has_tld
    except Exception:
        result = False

    return {"ok": True, "value": value, "result": result}
