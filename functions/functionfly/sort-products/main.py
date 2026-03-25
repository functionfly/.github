def handler(event):
    products = event.get("products") if isinstance(event, dict) else None
    sort_by = event.get("sort_by", "price")
    order = event.get("order", "asc")
    if not products:
        return {"ok": False, "error": "products is required"}
    try:
        reverse = order.lower() == "desc"
        valid_fields = {"price", "rating", "name", "popularity_score", "created_at", "discount_pct"}
        if sort_by not in valid_fields:
            return {"ok": False, "error": f"sort_by must be one of: {', '.join(sorted(valid_fields))}"}
        def sort_key(p):
            val = p.get(sort_by)
            if val is None:
                return (1, "") if sort_by == "name" else (float("inf"),)
            if sort_by == "name":
                return (0, str(val).lower())
            return (0, float(val))
        sorted_products = sorted(products, key=sort_key, reverse=reverse)
        return {"ok": True, "result": sorted_products, "count": len(sorted_products), "sorted_by": sort_by, "order": order}
    except Exception as e:
        return {"ok": False, "error": str(e)}
