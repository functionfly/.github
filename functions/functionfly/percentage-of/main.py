def handler(event):
    if isinstance(event, dict):
        v = event.get("value")
        p = event.get("percentage")
    else:
        v, p = None, None
    if v is None or p is None:
        return {"ok": False, "error": "value and percentage are required"}
    try:
        v, p = float(v), float(p)
        return {"ok": True, "result": round(v * p / 100, 10)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
