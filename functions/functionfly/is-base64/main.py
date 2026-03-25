import re
import base64


BASE64_RE = re.compile(r'^[A-Za-z0-9+/]*={0,2}$')
BASE64URL_RE = re.compile(r'^[A-Za-z0-9\-_]*={0,2}$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    url_safe = event.get("url_safe", False)

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()
    pattern = BASE64URL_RE if url_safe else BASE64_RE

    if not pattern.match(val):
        return {"ok": True, "value": value, "result": False}

    # Length must be a multiple of 4 (with padding)
    padded = val + '=' * ((4 - len(val) % 4) % 4)
    try:
        if url_safe:
            base64.urlsafe_b64decode(padded)
        else:
            base64.b64decode(padded)
        result = True
    except Exception:
        result = False

    return {"ok": True, "value": value, "result": result, "url_safe": url_safe}
