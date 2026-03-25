def handler(event):
    current_product = event.get("current_product") if isinstance(event, dict) else None
    available_products = event.get("available_products", [])
    max_suggestions = int(event.get("max_suggestions", 3))
    upsell_price_threshold_pct = float(event.get("upsell_price_threshold_pct", 50))
    if not current_product:
        return {"ok": False, "error": "current_product is required"}
    try:
        current_price = float(current_product.get("price", 0))
        current_category = str(current_product.get("category", "")).lower()
        current_id = str(current_product.get("id", ""))
        max_upsell_price = current_price * (1 + upsell_price_threshold_pct / 100)
        candidates = []
        for p in available_products:
            p_id = str(p.get("id", ""))
            if p_id == current_id:
                continue
            p_price = float(p.get("price", 0))
            p_category = str(p.get("category", "")).lower()
            if p_price > current_price and p_price <= max_upsell_price and p_category == current_category:
                score = float(p.get("rating", 3)) + (p_price - current_price) / current_price
                candidates.append({**p, "_upsell_score": round(score, 4), "_price_diff": round(p_price - current_price, 2)})
        suggestions = sorted(candidates, key=lambda x: x["_upsell_score"], reverse=True)[:max_suggestions]
        return {"ok": True, "result": suggestions, "count": len(suggestions), "current_product_id": current_id}
    except Exception as e:
        return {"ok": False, "error": str(e)}
