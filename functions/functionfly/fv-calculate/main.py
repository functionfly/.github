def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pv = event.get("present_value")
    rate = event.get("rate")
    periods = event.get("periods")
    if pv is None or rate is None or periods is None:
        return {"ok": False, "error": "present_value, rate, and periods are required"}
    try:
        pv = float(pv)
        rate = float(rate)
        periods = float(periods)
        pmt = float(event.get("payment", 0))
        if rate == 0:
            fv = pv + pmt * periods
        else:
            fv = pv * (1 + rate) ** periods + pmt * ((1 + rate) ** periods - 1) / rate
        return {"ok": True, "result": round(fv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
