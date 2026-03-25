def handler(event):
    r = event.get("r") if isinstance(event, dict) else None
    g, b = event.get("g"), event.get("b")
    if r is None or g is None or b is None:
        return {"ok": False, "error": "r, g, b are required"}
    try:
        r_, g_, b_ = int(r)/255, int(g)/255, int(b)/255
        cmax, cmin = max(r_, g_, b_), min(r_, g_, b_)
        delta = cmax - cmin
        v = cmax
        s = 0 if cmax == 0 else delta / cmax
        if delta == 0:
            h = 0
        elif cmax == r_:
            h = 60 * (((g_ - b_) / delta) % 6)
        elif cmax == g_:
            h = 60 * ((b_ - r_) / delta + 2)
        else:
            h = 60 * ((r_ - g_) / delta + 4)
        return {"ok": True, "result": {"h": round(h, 2), "s": round(s*100, 2), "v": round(v*100, 2)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
