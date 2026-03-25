TAX_CATEGORIES = {
    "standard": 1.0,
    "reduced": 0.5,
    "zero": 0.0,
    "exempt": 0.0,
    "luxury": 1.5,
    "food": 0.0,
    "digital": 1.0,
}


def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    base_rate = event.get("base_rate")
    product_category = event.get("product_category", "standard")
    jurisdiction = event.get("jurisdiction", "")
    quantity = int(event.get("quantity", 1))
    if price is None or base_rate is None:
        return {"ok": False, "error": "price and base_rate are required"}
    try:
        p, r = float(price), float(base_rate)
        multiplier = TAX_CATEGORIES.get(product_category.lower(), 1.0)
        effective_rate = round(r * multiplier, 4)
        unit_tax = round(p * effective_rate / 100, 2)
        total_tax = round(unit_tax * quantity, 2)
        total_price = round((p + unit_tax) * quantity, 2)
        return {
            "ok": True,
            "result": total_tax,
            "unit_price": p,
            "quantity": quantity,
            "base_rate_pct": r,
            "effective_rate_pct": effective_rate,
            "product_category": product_category,
            "unit_tax": unit_tax,
            "total_tax": total_tax,
            "total_price": total_price,
            "jurisdiction": jurisdiction
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
