from datetime import datetime
from zoneinfo import ZoneInfo


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date_string", event.get("date", ""))
        tz_name = event.get("timezone")
    else:
        date_str, tz_name = "", None

    if not date_str:
        return {"ok": False, "error": "Input 'date_string' is required"}

    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        if tz_name and dt.tzinfo is None:
            dt = dt.replace(tzinfo=ZoneInfo(tz_name))
        return {"ok": True, "iso": dt.isoformat()}
    except ValueError:
        pass
    for fmt in ("%Y-%m-%d", "%Y-%m-%d %H:%M:%S", "%d/%m/%Y", "%m/%d/%Y", "%B %d, %Y", "%b %d, %Y"):
        try:
            dt = datetime.strptime(date_str.strip(), fmt)
            if tz_name:
                dt = dt.replace(tzinfo=ZoneInfo(tz_name))
            return {"ok": True, "iso": dt.isoformat()}
        except ValueError:
            continue
    return {"ok": False, "error": "Could not parse date string"}
