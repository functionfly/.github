def handler(event):
    demand = event.get("demand") if isinstance(event, dict) else None
    stock = event.get("stock")
    if demand is None or stock is None:
        return {"ok": False, "error": "demand and stock are required"}
    try:
        d, s = int(demand), int(stock)
        backorder = max(0, d - s)
        fulfilled = min(d, s)
        fulfillment_rate = round(fulfilled / d * 100, 2) if d else 100
        return {
            "ok": True,
            "result": backorder,
            "backorder_qty": backorder,
            "fulfilled_qty": fulfilled,
            "unfulfilled_qty": backorder,
            "fulfillment_rate_pct": fulfillment_rate,
            "has_backorder": backorder > 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
