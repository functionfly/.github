def handler(event):
    product_a_orders = event.get("product_a_orders") if isinstance(event, dict) else None
    product_b_orders = event.get("product_b_orders")
    both_orders = event.get("both_orders")
    total_orders = event.get("total_orders")
    if product_a_orders is None or product_b_orders is None or both_orders is None or total_orders is None:
        return {"ok": False, "error": "product_a_orders, product_b_orders, both_orders, and total_orders are required"}
    try:
        a, b, ab, n = int(product_a_orders), int(product_b_orders), int(both_orders), int(total_orders)
        if n == 0:
            return {"ok": False, "error": "total_orders must be > 0"}
        support = ab / n
        confidence_a_to_b = ab / a if a > 0 else 0
        confidence_b_to_a = ab / b if b > 0 else 0
        expected = (a / n) * (b / n)
        lift = round(support / expected, 4) if expected > 0 else 0
        return {
            "ok": True,
            "result": lift,
            "lift": lift,
            "support": round(support, 4),
            "confidence_a_to_b": round(confidence_a_to_b, 4),
            "confidence_b_to_a": round(confidence_b_to_a, 4),
            "interpretation": "strong" if lift > 2 else ("positive" if lift > 1 else ("negative" if lift < 1 else "neutral"))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
