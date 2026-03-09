import calendar


def handler(event):
    if isinstance(event, dict):
        year = event.get("year")
    else:
        year = None
    if year is None:
        return {"ok": False, "error": "Input 'year' is required"}
    try:
        y = int(year)
        return {"ok": True, "is_leap_year": calendar.isleap(y)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
