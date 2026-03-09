from datetime import datetime

try:
    from zoneinfo import ZoneInfo
except ImportError:
    ZoneInfo = None  # Python < 3.9


def handler(event):
    """
    Convert timestamps between timezones.
    Uses zoneinfo (Python 3.9+) for IANA timezone names.

    Input:
        - timestamp: ISO datetime string or Unix timestamp (required)
        - from_tz: Source timezone (e.g. "UTC", "America/New_York")
        - to_tz: Target timezone (e.g. "Europe/London")

    Returns:
        - ok: True on success
        - iso: ISO format in target timezone
        - unix: Unix timestamp (seconds)
        - error: Message if ok is False
    """
    if ZoneInfo is None:
        return {"ok": False, "error": "Timezone support requires Python 3.9+ (zoneinfo)"}

    if isinstance(event, dict):
        ts = event.get("timestamp", event.get("ts", event.get("datetime")))
        from_tz = event.get("from_tz", event.get("from_timezone", "UTC"))
        to_tz = event.get("to_tz", event.get("to_timezone", "UTC"))
    else:
        ts = event
        from_tz = "UTC"
        to_tz = "UTC"

    if ts is None or ts == "":
        return {"ok": False, "error": "Input 'timestamp' is required"}

    try:
        from_zone = ZoneInfo(str(from_tz))
    except Exception as e:
        return {"ok": False, "error": f"Invalid from_tz: {e}"}
    try:
        to_zone = ZoneInfo(str(to_tz))
    except Exception as e:
        return {"ok": False, "error": f"Invalid to_tz: {e}"}

    dt = None
    if isinstance(ts, (int, float)):
        dt = datetime.fromtimestamp(float(ts), tz=from_zone)
    else:
        ts_str = str(ts).strip()
        if ts_str.isdigit():
            dt = datetime.fromtimestamp(int(ts_str), tz=from_zone)
        else:
            try:
                dt = datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=from_zone)
            except ValueError:
                return {"ok": False, "error": "Invalid timestamp format; use ISO 8601 or Unix seconds"}

    converted = dt.astimezone(to_zone)
    unix = int(converted.timestamp())
    iso = converted.isoformat()

    return {"ok": True, "iso": iso, "unix": unix, "timezone": to_tz}
