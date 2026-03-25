def handler(event):
    cart_total = event.get("cart_total") if isinstance(event, dict) else None
    breakpoints = event.get("breakpoints")
    if cart_total is None:
        return {"ok": False, "error": "cart_total is required"}
    try:
        total = float(cart_total)
        default_breakpoints = [
            {"min": 0, "max": 25, "tier": "bronze", "benefits": ["standard shipping"]},
            {"min": 25, "max": 50, "tier": "silver", "benefits": ["free standard shipping"]},
            {"min": 50, "max": 100, "tier": "gold", "benefits": ["free express shipping", "5% discount"]},
            {"min": 100, "max": None, "tier": "platinum", "benefits": ["free overnight shipping", "10% discount", "priority support"]},
        ]
        tiers = breakpoints if breakpoints else default_breakpoints
        current_tier = None
        next_tier = None
        for tier in sorted(tiers, key=lambda t: t.get("min", 0)):
            min_val = float(tier.get("min", 0))
            max_val = tier.get("max")
            if total >= min_val and (max_val is None or total < float(max_val)):
                current_tier = tier
            elif total < min_val and next_tier is None:
                next_tier = tier
        amount_to_next = round(float(next_tier["min"]) - total, 2) if next_tier else 0
        return {
            "ok": True,
            "result": current_tier.get("tier") if current_tier else None,
            "cart_total": total,
            "current_tier": current_tier,
            "next_tier": next_tier,
            "amount_to_next_tier": amount_to_next
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
