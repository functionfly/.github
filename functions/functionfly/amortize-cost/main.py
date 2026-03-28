def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    useful_life = event.get("useful_life")
    if cost is None or useful_life is None:
        return {"ok": False, "error": "cost and useful_life are required"}
    try:
        cost = float(cost)
        useful_life = float(useful_life)
        period = int(event.get("period", 1))
        if useful_life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        annual = cost / useful_life
        accumulated = annual * period
        book_value = max(cost - accumulated, 0)
        return {
            "ok": True,
            "result": round(annual, 2),
            "annual_amortization": round(annual, 2),
            "accumulated_amortization": round(accumulated, 2),
            "book_value": round(book_value, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
