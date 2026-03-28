def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    bv = event.get("beginning_value")
    ev = event.get("ending_value")
    years = event.get("years")
    if bv is None or ev is None or years is None:
        return {"ok": False, "error": "beginning_value, ending_value, and years are required"}
    try:
        bv = float(bv)
        ev = float(ev)
        years = float(years)
        if bv <= 0:
            return {"ok": False, "error": "beginning_value must be positive"}
        if years <= 0:
            return {"ok": False, "error": "years must be positive"}
        cagr = (ev / bv) ** (1 / years) - 1
        return {"ok": True, "result": round(cagr, 6), "result_pct": round(cagr * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
