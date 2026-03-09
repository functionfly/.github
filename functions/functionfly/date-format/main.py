from datetime import datetime
from zoneinfo import ZoneInfo


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date", "")
        fmt = event.get("format", "")
        tz_name = event.get("timezone")
    else:
        date_str, fmt, tz_name = "", "", None

    if not date_str or not fmt:
        return {"ok": False, "error": "Input 'date' and 'format' are required"}

    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        if tz_name:
            dt = dt.astimezone(ZoneInfo(tz_name))
        return {"ok": True, "formatted": dt.strftime(fmt)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
