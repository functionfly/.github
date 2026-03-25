def handler(event):
    query = event.get("query") if isinstance(event, dict) else None
    products = event.get("products", [])
    weights = event.get("weights", {"relevance": 0.5, "rating": 0.25, "popularity": 0.15, "price_score": 0.1})
    if not query:
        return {"ok": False, "error": "query is required"}
    try:
        q = str(query).lower()
        scored = []
        for p in products:
            name = str(p.get("name", "")).lower()
            desc = str(p.get("description", "")).lower()
            tags = " ".join(p.get("tags", [])).lower()
            text = f"{name} {desc} {tags}"
            words = q.split()
            matches = sum(1 for w in words if w in text)
            relevance = matches / max(len(words), 1)
            rating = float(p.get("rating", 3.0)) / 5.0
            popularity = float(p.get("popularity_score", 0.5))
            price = float(p.get("price", 100))
            price_score = max(0, 1 - price / 1000)
            w = weights
            score = round(
                relevance * w.get("relevance", 0.5) +
                rating * w.get("rating", 0.25) +
                popularity * w.get("popularity", 0.15) +
                price_score * w.get("price_score", 0.1), 4)
            scored.append({**p, "_score": score, "_relevance": round(relevance, 4)})
        ranked = sorted(scored, key=lambda x: x["_score"], reverse=True)
        return {"ok": True, "result": ranked, "count": len(ranked), "query": query}
    except Exception as e:
        return {"ok": False, "error": str(e)}
