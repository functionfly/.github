import re

BINARY_RE = re.compile(r'^(0b)?[01]+$', re.IGNORECASE)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    result = bool(BINARY_RE.match(val)) and len(val) > 0
    return {"ok": True, "value": value, "result": result}
