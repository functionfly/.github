def handler(event):
    stock_quantity = event.get("stock_quantity") if isinstance(event, dict) else None
    requested_quantity = int(event.get("requested_quantity", 1))
    low_stock_threshold = int(event.get("low_stock_threshold", 10))
    if stock_quantity is None:
        return {"ok": False, "error": "stock_quantity is required"}
    try:
        stock = int(stock_quantity)
        in_stock = stock > 0
        can_fulfill = stock >= requested_quantity
        low_stock = 0 < stock <= low_stock_threshold
        out_of_stock = stock == 0
        status = "out_of_stock" if out_of_stock else ("low_stock" if low_stock else "in_stock")
        return {
            "ok": True,
            "result": can_fulfill,
            "can_fulfill": can_fulfill,
            "in_stock": in_stock,
            "status": status,
            "available": stock,
            "requested": requested_quantity,
            "shortage": max(0, requested_quantity - stock)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
