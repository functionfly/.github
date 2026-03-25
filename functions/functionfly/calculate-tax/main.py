def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    rate = event.get("rate")
    if price is None or rate is None:
        return {"ok": False, "error": "price and rate are required"}
    try:
        p, r = float(price), float(rate)
        if p < 0:
            return {"ok": False, "error": "price must be >= 0"}
        if r < 0:
            return {"ok": False, "error": "rate must be >= 0"}
        tax = round(p * r / 100, 2)
        return {"ok": True, "result": tax, "tax": tax, "price": p, "rate": r, "total": round(p + tax, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
