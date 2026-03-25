def handler(event):
    wishlist_a = event.get("wishlist_a") if isinstance(event, dict) else None
    wishlist_b = event.get("wishlist_b")
    similarity_type = event.get("similarity_type", "jaccard")
    if not wishlist_a or not wishlist_b:
        return {"ok": False, "error": "wishlist_a and wishlist_b are required"}
    try:
        set_a = set(str(x) for x in wishlist_a)
        set_b = set(str(x) for x in wishlist_b)
        if similarity_type == "jaccard":
            intersection = len(set_a & set_b)
            union = len(set_a | set_b)
            similarity = round(intersection / union, 4) if union else 0
        elif similarity_type == "overlap":
            intersection = len(set_a & set_b)
            similarity = round(intersection / min(len(set_a), len(set_b)), 4) if min(len(set_a), len(set_b)) else 0
        else:
            return {"ok": False, "error": "similarity_type must be 'jaccard' or 'overlap'"}
        common = list(set_a & set_b)
        return {
            "ok": True,
            "result": similarity,
            "similarity": similarity,
            "similarity_type": similarity_type,
            "common_items": common,
            "common_count": len(common),
            "size_a": len(set_a),
            "size_b": len(set_b)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
