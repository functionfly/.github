def handler(event):
    quantity = event.get("quantity") if isinstance(event, dict) else None
    tiers = event.get("tiers")
    if quantity is None or not tiers:
        return {"ok": False, "error": "quantity and tiers are required"}
    try:
        qty = int(quantity)
        sorted_tiers = sorted(tiers, key=lambda t: t.get("min_qty", 0))
        price_per_unit = None
        applied_tier = None
        for tier in sorted_tiers:
            min_qty = int(tier.get("min_qty", 0))
            max_qty = tier.get("max_qty")
            if qty >= min_qty and (max_qty is None or qty <= int(max_qty)):
                price_per_unit = float(tier["price"])
                applied_tier = tier
        if price_per_unit is None:
            price_per_unit = float(sorted_tiers[0]["price"])
            applied_tier = sorted_tiers[0]
        total = round(price_per_unit * qty, 2)
        return {"ok": True, "result": price_per_unit, "price_per_unit": price_per_unit, "quantity": qty, "total": total, "applied_tier": applied_tier}
    except Exception as e:
        return {"ok": False, "error": str(e)}
