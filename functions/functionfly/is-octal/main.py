import re

OCTAL_RE = re.compile(r'^0?[0-7]+$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    result = bool(OCTAL_RE.match(val)) and len(val) > 0
    return {"ok": True, "value": value, "result": result}
