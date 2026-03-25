from email.utils import parsedate_to_datetime
from datetime import timezone


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    try:
        dt = parsedate_to_datetime(value.strip())
        dt_utc = dt.astimezone(timezone.utc)
        return {
            "ok": True,
            "value": value,
            "iso": dt_utc.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "timestamp": int(dt_utc.timestamp()),
        }
    except Exception as e:
        return {"ok": False, "error": f"Failed to parse Last-Modified header: {str(e)}"}
