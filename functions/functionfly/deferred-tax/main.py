def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    diff = event.get("temporary_difference")
    tax_rate = event.get("tax_rate")
    if diff is None or tax_rate is None:
        return {"ok": False, "error": "temporary_difference and tax_rate are required"}
    try:
        diff = float(diff)
        tax_rate = float(tax_rate)
        deferred = diff * tax_rate
        dtype = "liability" if diff > 0 else "asset"
        return {"ok": True, "result": round(deferred, 2), "type": dtype}
    except Exception as e:
        return {"ok": False, "error": str(e)}
