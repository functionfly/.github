def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    net_income = event.get("net_income")
    economic_capital = event.get("economic_capital")
    if net_income is None:
        return {"ok": False, "error": "net_income is required"}
    if economic_capital is None:
        return {"ok": False, "error": "economic_capital is required"}
    try:
        ni = float(net_income)
        ec = float(economic_capital)
        el = float(event.get("expected_loss", 0))
        if ec <= 0:
            return {"ok": False, "error": "economic_capital must be positive"}
        risk_adjusted_return = ni - el
        raroc = risk_adjusted_return / ec
        return {
            "ok": True,
            "result": round(raroc, 6),
            "raroc": round(raroc, 6),
            "raroc_pct": round(raroc * 100, 2),
            "net_income": ni,
            "economic_capital": ec,
            "expected_loss": el,
            "risk_adjusted_return": round(risk_adjusted_return, 2),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
