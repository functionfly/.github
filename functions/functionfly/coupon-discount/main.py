def handler(event):
    cart_total = event.get("cart_total") if isinstance(event, dict) else None
    discount_type = event.get("discount_type", "percent")
    discount_value = event.get("discount_value")
    max_discount = event.get("max_discount")
    min_cart = event.get("min_cart_total", 0)
    if cart_total is None or discount_value is None:
        return {"ok": False, "error": "cart_total and discount_value are required"}
    try:
        total = float(cart_total)
        val = float(discount_value)
        if total < float(min_cart):
            return {"ok": True, "result": 0, "discount_amount": 0, "final_total": total, "reason": "minimum cart total not met"}
        if discount_type == "percent":
            discount_amount = round(total * val / 100, 2)
        elif discount_type == "fixed":
            discount_amount = round(min(val, total), 2)
        elif discount_type == "free_shipping":
            return {"ok": True, "result": 0, "discount_amount": 0, "final_total": total, "free_shipping": True}
        else:
            return {"ok": False, "error": "discount_type must be 'percent', 'fixed', or 'free_shipping'"}
        if max_discount:
            discount_amount = min(discount_amount, float(max_discount))
        final_total = round(max(0, total - discount_amount), 2)
        return {"ok": True, "result": discount_amount, "discount_amount": discount_amount, "final_total": final_total, "original_total": total}
    except Exception as e:
        return {"ok": False, "error": str(e)}
