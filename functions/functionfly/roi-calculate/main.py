def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    gain = event.get("gain")
    cost = event.get("cost")
    if gain is None or cost is None:
        return {"ok": False, "error": "gain and cost are required"}
    try:
        gain = float(gain)
        cost = float(cost)
        if cost == 0:
            return {"ok": False, "error": "cost cannot be zero"}
        roi = gain / cost
        return {"ok": True, "result": round(roi, 6), "result_pct": round(roi * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
