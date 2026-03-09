from datetime import datetime
from zoneinfo import ZoneInfo


def handler(event):
    if isinstance(event, dict):
        ts = event.get("timestamp")
        tz_name = event.get("timezone", "UTC")
        fmt = event.get("format", "iso")
    else:
        ts, tz_name, fmt = None, "UTC", "iso"

    if ts is None:
        return {"ok": False, "error": "Input 'timestamp' is required"}
    try:
        if ts > 1e12:
            ts = ts / 1000.0
        dt = datetime.fromtimestamp(float(ts), tz=ZoneInfo(tz_name))
    except Exception as e:
        return {"ok": False, "error": str(e)}

    iso = dt.isoformat()
    if fmt == "iso" or not fmt:
        return {"ok": True, "date": iso, "iso": iso}
    try:
        return {"ok": True, "date": dt.strftime(fmt), "iso": iso}
    except Exception as e:
        return {"ok": False, "error": str(e)}
