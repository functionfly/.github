def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    beta = event.get("beta")
    mr = event.get("market_return")
    if pr is None or rfr is None or beta is None or mr is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, beta, and market_return are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        beta = float(beta)
        mr = float(mr)
        expected_return = rfr + beta * (mr - rfr)
        alpha = pr - expected_return
        return {"ok": True, "result": round(alpha, 6), "expected_return": round(expected_return, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
