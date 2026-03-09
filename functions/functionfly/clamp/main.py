def handler(event):
    if isinstance(event, dict):
        v = event.get("value")
        lo = event.get("min")
        hi = event.get("max")
    else:
        v, lo, hi = None, None, None
    if v is None or lo is None or hi is None:
        return {"ok": False, "error": "value, min, and max are required"}
    try:
        v, lo, hi = float(v), float(lo), float(hi)
        if lo > hi:
            return {"ok": False, "error": "min must be <= max"}
        return {"ok": True, "result": max(lo, min(hi, v))}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
