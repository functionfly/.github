def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    amount = event.get("amount")
    rate = event.get("rate")
    if amount is None or rate is None:
        return {"ok": False, "error": "amount and rate are required"}
    try:
        amount = float(amount)
        rate = float(rate)
        fee_pct = float(event.get("fee_pct", 0))
        converted = amount * rate
        fee = converted * fee_pct / 100
        net = converted - fee
        return {"ok": True, "result": round(net, 6), "converted": round(converted, 6), "fee": round(fee, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
