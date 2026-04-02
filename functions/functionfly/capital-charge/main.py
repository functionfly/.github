def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    exposure = event.get("exposure")
    risk_weight = event.get("risk_weight")
    if exposure is None:
        return {"ok": False, "error": "exposure is required"}
    if risk_weight is None:
        return {"ok": False, "error": "risk_weight is required"}
    try:
        exposure = float(exposure)
        risk_weight = float(risk_weight)
        capital_ratio = float(event.get("capital_ratio", 0.08))
        risk_weighted_assets = exposure * risk_weight
        capital_charge = risk_weighted_assets * capital_ratio
        return {
            "ok": True,
            "result": round(capital_charge, 2),
            "capital_charge": round(capital_charge, 2),
            "risk_weighted_assets": round(risk_weighted_assets, 2),
            "exposure": exposure,
            "risk_weight": risk_weight,
            "capital_ratio": capital_ratio,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
