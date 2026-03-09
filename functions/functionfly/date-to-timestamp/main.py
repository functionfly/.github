from datetime import datetime
from zoneinfo import ZoneInfo


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date", event.get("iso", ""))
        tz_name = event.get("timezone", "UTC")
    else:
        date_str, tz_name = "", "UTC"

    if not date_str:
        return {"ok": False, "error": "Input 'date' is required"}

    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=ZoneInfo(tz_name))
        ts = int(dt.timestamp())
        return {"ok": True, "timestamp": ts}
    except Exception as e:
        return {"ok": False, "error": str(e)}
