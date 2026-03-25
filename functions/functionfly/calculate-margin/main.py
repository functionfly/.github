def handler(event):
    revenue = event.get("revenue") if isinstance(event, dict) else None
    cost = event.get("cost")
    if revenue is None or cost is None:
        return {"ok": False, "error": "revenue and cost are required"}
    try:
        r, c = float(revenue), float(cost)
        if r == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        profit = round(r - c, 2)
        margin = round(profit / r * 100, 2)
        markup = round(profit / c * 100, 2) if c else None
        return {"ok": True, "result": margin, "revenue": r, "cost": c, "profit": profit, "margin_pct": margin, "markup_pct": markup}
    except Exception as e:
        return {"ok": False, "error": str(e)}
