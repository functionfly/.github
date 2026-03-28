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
        fv = float(event.get("future_value", 0))
        pmt = float(event.get("payment", 0))
        if rate == 0:
            pv = fv + pmt * periods
        else:
            pv = fv / (1 + rate) ** periods + pmt * (1 - (1 + rate) ** (-periods)) / rate
        return {"ok": True, "result": round(pv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
