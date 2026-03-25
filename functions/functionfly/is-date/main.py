from datetime import datetime


FORMATS = [
    "%Y-%m-%d",
    "%Y/%m/%d",
    "%d/%m/%Y",
    "%m/%d/%Y",
    "%d-%m-%Y",
    "%Y-%m-%dT%H:%M:%S",
    "%Y-%m-%dT%H:%M:%SZ",
    "%Y-%m-%dT%H:%M:%S.%f",
    "%Y-%m-%dT%H:%M:%S.%fZ",
    "%d %b %Y",
    "%d %B %Y",
    "%B %d, %Y",
]


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    fmt = event.get("format")

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()

    if fmt:
        try:
            dt = datetime.strptime(val, fmt)
            return {"ok": True, "value": value, "result": True, "format": fmt, "parsed": dt.isoformat()}
        except ValueError:
            return {"ok": True, "value": value, "result": False, "format": fmt}

    for f in FORMATS:
        try:
            dt = datetime.strptime(val, f)
            return {"ok": True, "value": value, "result": True, "format": f, "parsed": dt.isoformat()}
        except ValueError:
            continue

    return {"ok": True, "value": value, "result": False, "format": None}
