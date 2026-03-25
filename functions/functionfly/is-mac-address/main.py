import re

MAC_COLON = re.compile(r'^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$')
MAC_DASH = re.compile(r'^([0-9A-Fa-f]{2}-){5}[0-9A-Fa-f]{2}$')
MAC_DOT = re.compile(r'^([0-9A-Fa-f]{4}\.){2}[0-9A-Fa-f]{4}$')
MAC_NONE = re.compile(r'^[0-9A-Fa-f]{12}$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()
    if MAC_COLON.match(val):
        fmt = "colon"
    elif MAC_DASH.match(val):
        fmt = "dash"
    elif MAC_DOT.match(val):
        fmt = "dot"
    elif MAC_NONE.match(val):
        fmt = "none"
    else:
        return {"ok": True, "value": value, "result": False, "format": None}

    return {"ok": True, "value": value, "result": True, "format": fmt}
