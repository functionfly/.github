def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    holdings = event.get("holdings")
    if holdings is None:
        return {"ok": False, "error": "holdings is required"}
    try:
        total_weight = 0.0
        weighted_return = 0.0
        for h in holdings:
            w = float(h.get("weight", 0))
            r = float(h.get("return", 0))
            weighted_return += w * r
            total_weight += w
        if abs(total_weight - 1.0) > 0.01:
            return {"ok": False, "error": f"Weights must sum to 1.0, got {total_weight}"}
        return {"ok": True, "result": round(weighted_return, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
