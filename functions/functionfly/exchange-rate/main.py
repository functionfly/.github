def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    amount = event.get("amount")
    from_currency = event.get("from_currency")
    to_currency = event.get("to_currency")
    rate = event.get("rate")
    if amount is None or from_currency is None or to_currency is None or rate is None:
        return {"ok": False, "error": "amount, from_currency, to_currency, and rate are required"}
    try:
        amount = float(amount)
        rate = float(rate)
        converted = amount * rate
        return {
            "ok": True,
            "result": round(converted, 6),
            "from_currency": str(from_currency).upper(),
            "to_currency": str(to_currency).upper(),
            "rate": rate,
            "original_amount": amount
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
