import re

HEX_RE = re.compile(r'^(0x|0X)?[0-9a-fA-F]+$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    result = bool(HEX_RE.match(val))
    return {"ok": True, "value": value, "result": result}
