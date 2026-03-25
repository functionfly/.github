def handler(event):
    cart_total = event.get("cart_total") if isinstance(event, dict) else None
    threshold = float(event.get("threshold", 50.0))
    shipping_cost = float(event.get("shipping_cost", 5.99))
    if cart_total is None:
        return {"ok": False, "error": "cart_total is required"}
    try:
        total = float(cart_total)
        qualifies = total >= threshold
        remaining = round(max(0, threshold - total), 2)
        return {
            "ok": True,
            "result": qualifies,
            "qualifies_free_shipping": qualifies,
            "cart_total": total,
            "threshold": threshold,
            "remaining_for_free_shipping": remaining,
            "shipping_cost": 0 if qualifies else shipping_cost
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
