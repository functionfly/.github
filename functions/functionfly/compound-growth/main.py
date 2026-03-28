def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    rate = event.get("growth_rate")
    periods = event.get("periods")
    if initial is None or rate is None or periods is None:
        return {"ok": False, "error": "initial_value, growth_rate, and periods are required"}
    try:
        initial = float(initial)
        rate = float(rate)
        periods = float(periods)
        result = initial * (1 + rate) ** periods
        return {"ok": True, "result": round(result, 6), "growth": round(result - initial, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
