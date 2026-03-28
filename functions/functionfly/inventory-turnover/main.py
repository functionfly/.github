def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cogs = event.get("cost_of_goods_sold")
    avg_inv = event.get("average_inventory")
    if cogs is None or avg_inv is None:
        return {"ok": False, "error": "cost_of_goods_sold and average_inventory are required"}
    try:
        cogs = float(cogs)
        avg_inv = float(avg_inv)
        if avg_inv == 0:
            return {"ok": False, "error": "average_inventory cannot be zero"}
        turnover = cogs / avg_inv
        days = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_in_inventory": round(days, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
