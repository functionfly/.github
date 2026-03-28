def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    net_profit = event.get("net_profit")
    revenue = event.get("revenue")
    if net_profit is None or revenue is None:
        return {"ok": False, "error": "net_profit and revenue are required"}
    try:
        net_profit = float(net_profit)
        revenue = float(revenue)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        margin = net_profit / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
