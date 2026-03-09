from datetime import datetime, timedelta
from zoneinfo import ZoneInfo
import calendar


def _add_months(dt, months):
    m = dt.month - 1 + months
    y = dt.year + m // 12
    m = m % 12 + 1
    d = min(dt.day, calendar.monthrange(y, m)[1])
    return dt.replace(year=y, month=m, day=d)


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date", "")
        amount = event.get("amount", 0)
        unit = (event.get("unit") or "days").lower()
    else:
        date_str, amount, unit = "", 0, "days"

    if not date_str:
        return {"ok": False, "error": "Input 'date' is required"}
    try:
        amount = int(amount)
    except (TypeError, ValueError):
        return {"ok": False, "error": "Input 'amount' must be an integer"}

    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
    except Exception as e:
        return {"ok": False, "error": str(e)}

    try:
        if unit in ("days", "day", "d"):
            result = dt + timedelta(days=amount)
        elif unit in ("weeks", "week", "w"):
            result = dt + timedelta(weeks=amount)
        elif unit in ("hours", "hour", "h"):
            result = dt + timedelta(hours=amount)
        elif unit in ("minutes", "minute", "m", "mins"):
            result = dt + timedelta(minutes=amount)
        elif unit in ("seconds", "second", "s", "secs"):
            result = dt + timedelta(seconds=amount)
        elif unit in ("months", "month"):
            result = _add_months(dt, amount)
        elif unit in ("years", "year"):
            result = _add_months(dt, amount * 12)
        else:
            return {"ok": False, "error": f"Unknown unit: {unit}"}
        iso = result.isoformat()
        return {"ok": True, "result": iso, "iso": iso}
    except Exception as e:
        return {"ok": False, "error": str(e)}
