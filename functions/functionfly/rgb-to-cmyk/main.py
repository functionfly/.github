def handler(event):
    r = event.get("r") if isinstance(event, dict) else None
    g, b = event.get("g"), event.get("b")
    if r is None or g is None or b is None:
        return {"ok": False, "error": "r, g, b are required"}
    try:
        r_, g_, b_ = int(r)/255, int(g)/255, int(b)/255
        k = 1 - max(r_, g_, b_)
        if k == 1:
            return {"ok": True, "result": {"c": 0, "m": 0, "y": 0, "k": 100}}
        c = round((1 - r_ - k) / (1 - k) * 100, 2)
        m = round((1 - g_ - k) / (1 - k) * 100, 2)
        y = round((1 - b_ - k) / (1 - k) * 100, 2)
        return {"ok": True, "result": {"c": c, "m": m, "y": y, "k": round(k * 100, 2)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
