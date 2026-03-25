import re

SPECIAL_RE = re.compile(r'[^a-zA-Z0-9\s]')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value)
    matches = SPECIAL_RE.findall(val)
    result = len(matches) > 0
    return {"ok": True, "value": value, "result": result, "special_chars": list(set(matches))}
