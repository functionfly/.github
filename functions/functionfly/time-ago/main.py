from datetime import datetime, timezone

def _human(delta_sec):
    if delta_sec < 60: return "just now"
    if delta_sec < 3600: return f"{int(delta_sec/60)} minute(s) ago"
    if delta_sec < 86400: return f"{int(delta_sec/3600)} hour(s) ago"
    if delta_sec < 2592000: return f"{int(delta_sec/86400)} day(s) ago"
    if delta_sec < 31536000: return f"{int(delta_sec/2592000)} month(s) ago"
    return f"{int(delta_sec/31536000)} year(s) ago"

def handler(event):
    date_str = event.get("date", "") if isinstance(event, dict) else ""
    base_str = event.get("base") if isinstance(event, dict) else None
    if not date_str:
        return {"ok": False, "error": "Input date is required"}
    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        if dt.tzinfo is None: dt = dt.replace(tzinfo=timezone.utc)
        base = datetime.fromisoformat(base_str.replace("Z", "+00:00")) if base_str else datetime.now(timezone.utc)
        if base.tzinfo is None: base = base.replace(tzinfo=timezone.utc)
        delta_sec = (base - dt).total_seconds()
        if delta_sec < 0: return {"ok": False, "error": "date is in the future"}
        return {"ok": True, "text": _human(delta_sec)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
