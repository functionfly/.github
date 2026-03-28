def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    cash_flows = event.get("cash_flows")
    if rate is None or cash_flows is None:
        return {"ok": False, "error": "rate and cash_flows are required"}
    try:
        rate = float(rate)
        cash_flows = [float(cf) for cf in cash_flows]
        npv = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
        return {"ok": True, "result": round(npv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
