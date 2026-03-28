def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = sorted([float(x) for x in returns])
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        confidence = float(event.get("confidence_level", 0.95))
        portfolio_value = float(event.get("portfolio_value", 1))
        cutoff = int((1 - confidence) * n)
        if cutoff == 0:
            cutoff = 1
        tail = r[:cutoff]
        cvar = (sum(tail) / len(tail)) * portfolio_value
        return {"ok": True, "result": round(cvar, 8), "confidence_level": confidence}
    except Exception as e:
        return {"ok": False, "error": str(e)}
