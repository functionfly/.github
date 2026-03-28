def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    if rate is None or periods is None:
        return {"ok": False, "error": "rate and periods are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        if rate == 0:
            factor = periods
        else:
            factor = (1 - (1 + rate) ** (-periods)) / rate
        return {"ok": True, "result": round(factor, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
