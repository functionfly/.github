def handler(event):
    """
    Convert amounts between currencies using a given exchange rate.
    For live rates you would call an external API; here we accept rate as input.

    Input:
        - amount: Numeric amount in source currency (required)
        - from_currency: Source currency code (e.g. USD)
        - to_currency: Target currency code (e.g. EUR)
        - rate: Exchange rate (1 from_currency = rate to_currency). If omitted, only returns amount unchanged with a note.

    Returns:
        - ok: True on success
        - amount: Converted amount
        - from_currency, to_currency, rate_used
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        amount = event.get("amount", event.get("value", 0))
        from_currency = event.get("from_currency", event.get("from", "USD"))
        to_currency = event.get("to_currency", event.get("to", "EUR"))
        rate = event.get("rate")
    else:
        amount = event
        from_currency = "USD"
        to_currency = "EUR"
        rate = None

    try:
        amount = float(amount)
    except (TypeError, ValueError):
        return {"ok": False, "error": "Input 'amount' must be a number"}

    if rate is not None:
        try:
            rate = float(rate)
        except (TypeError, ValueError):
            return {"ok": False, "error": "Invalid 'rate'; must be a number"}
        converted = amount * rate
        return {
            "ok": True,
            "amount": round(converted, 2),
            "original_amount": amount,
            "from_currency": str(from_currency),
            "to_currency": str(to_currency),
            "rate_used": rate,
        }

    # No rate provided: return amount unchanged and suggest providing rate
    return {
        "ok": True,
        "amount": amount,
        "original_amount": amount,
        "from_currency": str(from_currency),
        "to_currency": str(to_currency),
        "rate_used": None,
        "message": "No rate provided; pass 'rate' for conversion (e.g. 1 USD = 0.92 EUR).",
    }
