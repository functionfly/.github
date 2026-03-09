import calendar


def handler(event):
    if isinstance(event, dict):
        year = event.get("year")
        month = event.get("month")
    else:
        year, month = None, None
    if year is None or month is None:
        return {"ok": False, "error": "Input 'year' and 'month' are required"}
    try:
        y, m = int(year), int(month)
        if not (1 <= m <= 12):
            return {"ok": False, "error": "month must be 1-12"}
        return {"ok": True, "days": calendar.monthrange(y, m)[1]}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
