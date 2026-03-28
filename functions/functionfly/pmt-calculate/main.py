def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    pv = event.get("present_value")
    if rate is None or periods is None or pv is None:
        return {"ok": False, "error": "rate, periods, and present_value are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        if rate == 0:
            pmt = -(pv + fv) / periods
        else:
            pmt = -rate * (pv * (1 + rate) ** periods + fv) / ((1 + rate) ** periods - 1)
        return {"ok": True, "result": round(pmt, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
