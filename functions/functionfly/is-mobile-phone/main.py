import re

MOBILE_RE = re.compile(r'^\+?[1-9]\d{6,14}$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()
    # Strip common formatting
    normalized = re.sub(r'[\s\-\.\(\)]', '', val)
    result = bool(MOBILE_RE.match(normalized))
    return {"ok": True, "value": value, "result": result, "normalized": normalized}
