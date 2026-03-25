def handler(event):
    products = event.get("products") if isinstance(event, dict) else None
    filters = event.get("filters", {})
    if not products:
        return {"ok": False, "error": "products is required"}
    try:
        result = list(products)
        if "min_price" in filters:
            result = [p for p in result if float(p.get("price", 0)) >= float(filters["min_price"])]
        if "max_price" in filters:
            result = [p for p in result if float(p.get("price", 0)) <= float(filters["max_price"])]
        if "min_rating" in filters:
            result = [p for p in result if float(p.get("rating", 0)) >= float(filters["min_rating"])]
        if "category" in filters:
            cats = filters["category"] if isinstance(filters["category"], list) else [filters["category"]]
            result = [p for p in result if str(p.get("category", "")).lower() in [c.lower() for c in cats]]
        if "tags" in filters:
            required_tags = set(t.lower() for t in filters["tags"])
            result = [p for p in result if required_tags.issubset(set(t.lower() for t in p.get("tags", [])))]
        if "in_stock" in filters and filters["in_stock"]:
            result = [p for p in result if p.get("stock_quantity", 0) > 0]
        if "brand" in filters:
            brands = filters["brand"] if isinstance(filters["brand"], list) else [filters["brand"]]
            result = [p for p in result if str(p.get("brand", "")).lower() in [b.lower() for b in brands]]
        return {"ok": True, "result": result, "count": len(result), "original_count": len(products), "filters_applied": list(filters.keys())}
    except Exception as e:
        return {"ok": False, "error": str(e)}
