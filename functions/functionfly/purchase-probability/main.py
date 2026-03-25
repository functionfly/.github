import math


def handler(event):
    page_views = int(event.get("page_views", 0)) if isinstance(event, dict) else 0
    time_on_page_seconds = float(event.get("time_on_page_seconds", 0))
    cart_add = event.get("cart_add", False)
    wishlist_add = event.get("wishlist_add", False)
    returning_customer = event.get("returning_customer", False)
    price = float(event.get("price", 50))
    avg_session_conversion_rate = float(event.get("avg_conversion_rate", 0.02))
    try:
        base = avg_session_conversion_rate
        if page_views > 1:
            base *= min(3.0, 1 + math.log(page_views))
        if time_on_page_seconds > 60:
            base *= min(2.5, 1 + time_on_page_seconds / 300)
        if cart_add:
            base *= 6.0
        if wishlist_add:
            base *= 2.0
        if returning_customer:
            base *= 1.5
        if price < 25:
            base *= 1.5
        elif price > 200:
            base *= 0.7
        probability = round(min(0.99, base), 4)
        return {
            "ok": True,
            "result": probability,
            "purchase_probability": probability,
            "confidence": "high" if probability > 0.4 else ("medium" if probability > 0.1 else "low"),
            "factors": {"cart_add": cart_add, "wishlist_add": wishlist_add, "page_views": page_views}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
