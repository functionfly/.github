def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    revenue = event.get("revenue")
    cogs = event.get("cost_of_goods_sold")
    if revenue is None or cogs is None:
        return {"ok": False, "error": "revenue and cost_of_goods_sold are required"}
    try:
        revenue = float(revenue)
        cogs = float(cogs)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        gross_profit = revenue - cogs
        margin = gross_profit / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4), "gross_profit": round(gross_profit, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
