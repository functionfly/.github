def handler(event):
    product = event.get("product") if isinstance(event, dict) else None
    customer_history = event.get("customer_history", [])
    weights = event.get("weights", {"category_match": 0.4, "rating": 0.3, "popularity": 0.2, "recency": 0.1})
    if not product:
        return {"ok": False, "error": "product is required"}
    try:
        p_category = str(product.get("category", "")).lower()
        p_rating = float(product.get("rating", 3.0))
        p_popularity = float(product.get("popularity_score", 0.5))
        p_days_old = float(product.get("days_old", 365))
        # Category match score
        history_categories = [str(h.get("category", "")).lower() for h in customer_history]
        category_match = (history_categories.count(p_category) / max(len(history_categories), 1)) if history_categories else 0.5
        # Normalize scores to 0-1
        rating_score = (p_rating - 1) / 4.0
        popularity_score = min(1.0, p_popularity)
        recency_score = max(0, 1 - p_days_old / 365)
        w = weights
        score = round(
            category_match * w.get("category_match", 0.4) +
            rating_score * w.get("rating", 0.3) +
            popularity_score * w.get("popularity", 0.2) +
            recency_score * w.get("recency", 0.1), 4)
        return {
            "ok": True,
            "result": score,
            "recommendation_score": score,
            "component_scores": {"category_match": round(category_match, 4), "rating": round(rating_score, 4), "popularity": round(popularity_score, 4), "recency": round(recency_score, 4)}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
