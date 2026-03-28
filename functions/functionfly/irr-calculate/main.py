def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash_flows = event.get("cash_flows")
    if cash_flows is None:
        return {"ok": False, "error": "cash_flows is required"}
    try:
        cash_flows = [float(cf) for cf in cash_flows]
        guess = float(event.get("guess", 0.1))
        # Newton-Raphson method
        rate = guess
        for _ in range(1000):
            npv = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
            dnpv = sum(-t * cf / (1 + rate) ** (t + 1) for t, cf in enumerate(cash_flows))
            if abs(dnpv) < 1e-12:
                break
            new_rate = rate - npv / dnpv
            if abs(new_rate - rate) < 1e-10:
                rate = new_rate
                break
            rate = new_rate
        npv_check = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
        if abs(npv_check) > 0.01:
            return {"ok": False, "error": "IRR did not converge"}
        return {"ok": True, "result": round(rate, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
