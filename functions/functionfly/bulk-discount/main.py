def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    quantity = event.get("quantity")
    thresholds = event.get("thresholds")
    if price is None or quantity is None:
        return {"ok": False, "error": "price and quantity are required"}
    try:
        p, qty = float(price), int(quantity)
        default_thresholds = [
            {"min_qty": 10, "discount_pct": 5},
            {"min_qty": 25, "discount_pct": 10},
            {"min_qty": 50, "discount_pct": 15},
            {"min_qty": 100, "discount_pct": 20},
        ]
        tiers = thresholds if thresholds else default_thresholds
        discount_pct = 0
        applied_tier = None
        for tier in sorted(tiers, key=lambda t: t.get("min_qty", 0)):
            if qty >= int(tier.get("min_qty", 0)):
                discount_pct = float(tier.get("discount_pct", 0))
                applied_tier = tier
        discounted_price = round(p * (1 - discount_pct / 100), 2)
        total = round(discounted_price * qty, 2)
        savings = round((p - discounted_price) * qty, 2)
        return {"ok": True, "result": discount_pct, "discount_pct": discount_pct, "unit_price": p, "discounted_price": discounted_price, "quantity": qty, "total": total, "savings": savings, "applied_tier": applied_tier}
    except Exception as e:
        return {"ok": False, "error": str(e)}
