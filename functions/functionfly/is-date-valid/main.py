from datetime import date


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "input must be an object"}

    value = event.get("value")
    year = event.get("year")
    month = event.get("month")
    day = event.get("day")

    try:
        if value is not None:
            parsed = date.fromisoformat(str(value))
            year, month, day = parsed.year, parsed.month, parsed.day
        elif year is not None and month is not None and day is not None:
            parsed = date(int(year), int(month), int(day))
            year, month, day = parsed.year, parsed.month, parsed.day
        else:
            return {"ok": False, "error": "provide value or year/month/day"}
        return {"ok": True, "result": True, "year": year, "month": month, "day": day}
    except (ValueError, TypeError):
        return {"ok": True, "result": False, "year": year, "month": month, "day": day}
