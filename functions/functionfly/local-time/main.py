from datetime import datetime
from zoneinfo import ZoneInfo

def handler(event):
    if isinstance(event, dict):
        utc_str = event.get("utc", "")
        tz_name = event.get("timezone", "")
    else:
        utc_str, tz_name = "", ""
    if not utc_str or not tz_name:
        return {"ok": False, "error": "Input utc and timezone are required"}
    try:
        dt = datetime.fromisoformat(utc_str.replace("Z", "+00:00"))
        if dt.tzinfo is None:
            from datetime import timezone
            dt = dt.replace(tzinfo=timezone.utc)
        local = dt.astimezone(ZoneInfo(tz_name))
        return {"ok": True, "local_iso": local.isoformat()}
    except Exception as e:
        return {"ok": False, "error": str(e)}
