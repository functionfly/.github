from datetime import datetime, timedelta


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date", "")
        n = event.get("days", 0)
    else:
        date_str, n = "", 0
    if not date_str:
        return {"ok": False, "error": "Input 'date' is required"}
    try:
        n = int(n)
    except (TypeError, ValueError):
        return {"ok": False, "error": "Input 'days' must be an integer"}
    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        if dt.tzinfo:
            dt = dt.replace(tzinfo=None)
        added = 0
        step = 1 if n >= 0 else -1
        while added != n:
            dt += timedelta(days=step)
            if dt.weekday() < 5:
                added += step
        iso = dt.isoformat()
        return {"ok": True, "result": iso, "iso": iso}
    except Exception as e:
        return {"ok": False, "error": str(e)}
