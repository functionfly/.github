def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    beta = event.get("beta")
    if pr is None or rfr is None or beta is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and beta are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        beta = float(beta)
        if beta == 0:
            return {"ok": False, "error": "beta cannot be zero"}
        treynor = (pr - rfr) / beta
        return {"ok": True, "result": round(treynor, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
