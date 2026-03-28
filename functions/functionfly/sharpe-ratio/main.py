def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    std = event.get("portfolio_std_dev")
    if pr is None or rfr is None or std is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and portfolio_std_dev are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        std = float(std)
        if std == 0:
            return {"ok": False, "error": "portfolio_std_dev cannot be zero"}
        sharpe = (pr - rfr) / std
        return {"ok": True, "result": round(sharpe, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
