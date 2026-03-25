def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    bundle_discount_pct = float(event.get("bundle_discount_pct", 10))
    if not items:
        return {"ok": False, "error": "items is required (list of {price, quantity})"}
    try:
        subtotal = sum(float(item["price"]) * int(item.get("quantity", 1)) for item in items)
        discount = round(subtotal * bundle_discount_pct / 100, 2)
        bundle_price = round(subtotal - discount, 2)
        return {
            "ok": True,
            "result": bundle_price,
            "subtotal": round(subtotal, 2),
            "bundle_discount": discount,
            "bundle_price": bundle_price,
            "item_count": len(items),
            "savings_pct": bundle_discount_pct
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
