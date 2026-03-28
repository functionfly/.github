def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    salvage = event.get("salvage_value")
    life = event.get("useful_life")
    period = event.get("period")
    if cost is None or salvage is None or life is None or period is None:
        return {"ok": False, "error": "cost, salvage_value, useful_life, and period are required"}
    try:
        cost = float(cost)
        salvage = float(salvage)
        life = int(life)
        period = int(period)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        if period < 1 or period > life:
            return {"ok": False, "error": f"period must be between 1 and {life}"}
        sum_of_years = life * (life + 1) / 2
        fraction = (life - period + 1) / sum_of_years
        dep = (cost - salvage) * fraction
        return {"ok": True, "result": round(dep, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
