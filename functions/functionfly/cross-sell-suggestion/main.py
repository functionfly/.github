def handler(event):
    cart_items = event.get("cart_items") if isinstance(event, dict) else None
    available_products = event.get("available_products", [])
    affinity_rules = event.get("affinity_rules", [])
    max_suggestions = int(event.get("max_suggestions", 3))
    if not cart_items:
        return {"ok": False, "error": "cart_items is required"}
    try:
        cart_ids = set(str(item.get("id", "")) for item in cart_items)
        cart_categories = set(str(item.get("category", "")).lower() for item in cart_items)
        suggestions = []
        for p in available_products:
            p_id = str(p.get("id", ""))
            if p_id in cart_ids:
                continue
            p_category = str(p.get("category", "")).lower()
            score = 0.0
            # Not same category = potential cross-sell
            if p_category not in cart_categories:
                score += 0.5
            # Check affinity rules
            for rule in affinity_rules:
                trigger_cats = set(c.lower() for c in rule.get("trigger_categories", []))
                suggest_cats = set(c.lower() for c in rule.get("suggest_categories", []))
                if trigger_cats.issubset(cart_categories) and p_category in suggest_cats:
                    score += float(rule.get("weight", 1.0))
            score += float(p.get("rating", 3)) / 5.0 * 0.3
            if score > 0:
                suggestions.append({**p, "_cross_sell_score": round(score, 4)})
        ranked = sorted(suggestions, key=lambda x: x["_cross_sell_score"], reverse=True)[:max_suggestions]
        return {"ok": True, "result": ranked, "count": len(ranked), "cart_categories": list(cart_categories)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
