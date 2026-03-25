def _hsl_to_rgb(h, s, l):
    s_, l_ = s / 100, l / 100
    c = (1 - abs(2 * l_ - 1)) * s_
    x = c * (1 - abs((h / 60) % 2 - 1))
    m = l_ - c / 2
    if 0 <= h < 60:   r_, g_, b_ = c, x, 0
    elif 60 <= h < 120: r_, g_, b_ = x, c, 0
    elif 120 <= h < 180: r_, g_, b_ = 0, c, x
    elif 180 <= h < 240: r_, g_, b_ = 0, x, c
    elif 240 <= h < 300: r_, g_, b_ = x, 0, c
    else:              r_, g_, b_ = c, 0, x
    return round((r_ + m) * 255), round((g_ + m) * 255), round((b_ + m) * 255)


def handler(event):
    h = event.get("h") if isinstance(event, dict) else None
    s, l = event.get("s"), event.get("l")
    if h is None or s is None or l is None:
        return {"ok": False, "error": "h, s, l are required"}
    try:
        r, g, b = _hsl_to_rgb(float(h) % 360, float(s), float(l))
        return {"ok": True, "result": {"r": r, "g": g, "b": b}, "css": f"rgb({r},{g},{b})", "hex": f"#{r:02X}{g:02X}{b:02X}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
