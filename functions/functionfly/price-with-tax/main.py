def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    rate = event.get("rate")
    if price is None or rate is None:
        return {"ok": False, "error": "price and rate are required"}
    try:
        p, r = float(price), float(rate)
        total = round(p * (1 + r / 100), 2)
        return {"ok": True, "result": total, "net": p, "tax": round(total - p, 2), "total": total}
    except Exception as e:
        return {"ok": False, "error": str(e)}
