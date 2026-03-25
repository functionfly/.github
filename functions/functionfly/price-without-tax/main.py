def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    rate = event.get("rate")
    if price is None or rate is None:
        return {"ok": False, "error": "price and rate are required"}
    try:
        p, r = float(price), float(rate)
        net = round(p / (1 + r / 100), 2)
        return {"ok": True, "result": net, "gross": p, "net": net, "tax": round(p - net, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
