def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    purchase_price = event.get("purchase_price")
    if purchase_price is None:
        return {"ok": False, "error": "purchase_price is required"}
    try:
        total = float(purchase_price)
        additional = event.get("additional_costs", [])
        for cost in additional:
            total += float(cost)
        return {"ok": True, "result": round(total, 2), "capitalized_cost": round(total, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
