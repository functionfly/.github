import re

DOMAIN_RE = re.compile(
    r'^(?!-)[A-Za-z0-9\-]{1,63}(?<!-)(\.[A-Za-z0-9\-]{1,63}(?<!-))*\.[A-Za-z]{2,}$'
)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).lower().strip()
    result = bool(DOMAIN_RE.match(val)) and len(val) <= 253
    return {"ok": True, "value": value, "result": result}
