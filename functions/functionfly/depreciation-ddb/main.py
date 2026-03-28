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
        life = float(life)
        period = int(period)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        if period < 1 or period > life:
            return {"ok": False, "error": f"period must be between 1 and {int(life)}"}
        rate = 2 / life
        book_value = cost
        dep = 0
        for p in range(1, period + 1):
            dep = min(rate * book_value, book_value - salvage)
            dep = max(dep, 0)
            book_value -= dep
        return {"ok": True, "result": round(dep, 6), "book_value_after": round(book_value, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
