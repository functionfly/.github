def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    dd = event.get("downside_deviation")
    if pr is None or rfr is None or dd is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and downside_deviation are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        dd = float(dd)
        if dd == 0:
            return {"ok": False, "error": "downside_deviation cannot be zero"}
        sortino = (pr - rfr) / dd
        return {"ok": True, "result": round(sortino, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
