def handler(event):
    h = event.get("h") if isinstance(event, dict) else None
    s, v = event.get("s"), event.get("v")
    if h is None or s is None or v is None:
        return {"ok": False, "error": "h, s, v are required"}
    try:
        h_, s_, v_ = float(h) % 360, float(s) / 100, float(v) / 100
        c = v_ * s_
        x = c * (1 - abs((h_ / 60) % 2 - 1))
        m = v_ - c
        if 0 <= h_ < 60:    r_, g_, b_ = c, x, 0
        elif 60 <= h_ < 120: r_, g_, b_ = x, c, 0
        elif 120 <= h_ < 180: r_, g_, b_ = 0, c, x
        elif 180 <= h_ < 240: r_, g_, b_ = 0, x, c
        elif 240 <= h_ < 300: r_, g_, b_ = x, 0, c
        else:                r_, g_, b_ = c, 0, x
        r, g, b = round((r_+m)*255), round((g_+m)*255), round((b_+m)*255)
        return {"ok": True, "result": {"r": r, "g": g, "b": b}, "hex": f"#{r:02X}{g:02X}{b:02X}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
