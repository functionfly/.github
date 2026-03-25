def handler(event):
    year = event.get("year") if isinstance(event, dict) else None
    if year is None:
        return {"ok": False, "error": "year is required"}
    try:
        y = int(year)
    except (TypeError, ValueError):
        return {"ok": False, "error": "year must be an integer"}
    result = (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)
    return {"ok": True, "year": y, "result": result}
