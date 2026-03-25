import re

HEX_COLOR_RE = re.compile(r'^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{4}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    match = HEX_COLOR_RE.match(val)
    result = bool(match)
    length = len(match.group(1)) if match else None
    has_alpha = length in (4, 8) if length else False
    return {"ok": True, "value": value, "result": result, "has_alpha": has_alpha}
