def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["equity_value", "debt_value", "cost_of_equity", "cost_of_debt", "tax_rate"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        E = float(event["equity_value"])
        D = float(event["debt_value"])
        Re = float(event["cost_of_equity"])
        Rd = float(event["cost_of_debt"])
        T = float(event["tax_rate"])
        V = E + D
        if V == 0:
            return {"ok": False, "error": "Total value (equity + debt) cannot be zero"}
        wacc = (E / V) * Re + (D / V) * Rd * (1 - T)
        return {"ok": True, "result": round(wacc, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
