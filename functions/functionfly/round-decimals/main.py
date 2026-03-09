def handler(event):
    if isinstance(event, dict):
        v = event.get("value")
        d = event.get("decimals", 0)
    else:
        v, d = None, 0
    if v is None:
        return {"ok": False, "error": "value is required"}
    try:
        v = float(v)
        d = int(d)
        return {"ok": True, "result": round(v, d)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
