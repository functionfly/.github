from datetime import datetime


def handler(event):
    if isinstance(event, dict):
        a = event.get("date_a", "")
        b = event.get("date_b", "")
        unit = (event.get("unit") or "days").lower()
    else:
        a, b, unit = "", "", "days"

    if not a or not b:
        return {"ok": False, "error": "Input 'date_a' and 'date_b' are required"}
    try:
        dt_a = datetime.fromisoformat(a.replace("Z", "+00:00"))
        dt_b = datetime.fromisoformat(b.replace("Z", "+00:00"))
    except Exception as e:
        return {"ok": False, "error": str(e)}

    delta = dt_b - dt_a
    if unit in ("days", "day", "d"):
        diff = delta.total_seconds() / 86400
    elif unit in ("hours", "hour", "h"):
        diff = delta.total_seconds() / 3600
    elif unit in ("minutes", "minute", "m"):
        diff = delta.total_seconds() / 60
    elif unit in ("seconds", "second", "s"):
        diff = delta.total_seconds()
    else:
        return {"ok": False, "error": f"Unknown unit: {unit}"}
    return {"ok": True, "difference": round(diff, 6), "unit": unit}
