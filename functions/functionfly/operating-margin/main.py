def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    operating_income = event.get("operating_income")
    revenue = event.get("revenue")
    if operating_income is None or revenue is None:
        return {"ok": False, "error": "operating_income and revenue are required"}
    try:
        operating_income = float(operating_income)
        revenue = float(revenue)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        margin = operating_income / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
