def handler(event):
    cost = event.get("cost") if isinstance(event, dict) else None
    markup = event.get("markup")
    markup_type = event.get("type", "percent")
    if cost is None or markup is None:
        return {"ok": False, "error": "cost and markup are required"}
    try:
        c, m = float(cost), float(markup)
        if markup_type == "percent":
            price = round(c * (1 + m / 100), 2)
        else:
            price = round(c + m, 2)
        profit = round(price - c, 2)
        margin = round(profit / price * 100, 2) if price else 0
        return {"ok": True, "result": price, "cost": c, "markup": m, "selling_price": price, "profit": profit, "margin_pct": margin}
    except Exception as e:
        return {"ok": False, "error": str(e)}
