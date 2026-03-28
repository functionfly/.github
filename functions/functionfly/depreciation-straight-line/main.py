def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    salvage = event.get("salvage_value")
    life = event.get("useful_life")
    if cost is None or salvage is None or life is None:
        return {"ok": False, "error": "cost, salvage_value, and useful_life are required"}
    try:
        cost = float(cost)
        salvage = float(salvage)
        life = float(life)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        annual_dep = (cost - salvage) / life
        schedule = [{"year": y + 1, "depreciation": round(annual_dep, 2), "book_value": round(cost - annual_dep * (y + 1), 2)} for y in range(int(life))]
        return {"ok": True, "result": round(annual_dep, 6), "annual_depreciation": round(annual_dep, 2), "schedule": schedule}
    except Exception as e:
        return {"ok": False, "error": str(e)}
