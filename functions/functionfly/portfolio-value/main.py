def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    holdings = event.get("holdings")
    if holdings is None:
        return {"ok": False, "error": "holdings is required"}
    try:
        total = 0.0
        details = []
        for h in holdings:
            shares = float(h.get("shares", 0))
            price = float(h.get("price", 0))
            value = shares * price
            total += value
            details.append({"shares": shares, "price": price, "value": round(value, 2)})
        return {"ok": True, "result": round(total, 2), "holdings": details}
    except Exception as e:
        return {"ok": False, "error": str(e)}
