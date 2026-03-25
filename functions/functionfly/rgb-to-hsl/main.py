def _rgb_to_hsl(r, g, b):
    r_, g_, b_ = r / 255, g / 255, b / 255
    cmax, cmin = max(r_, g_, b_), min(r_, g_, b_)
    delta = cmax - cmin
    l = (cmax + cmin) / 2
    s = 0 if delta == 0 else delta / (1 - abs(2 * l - 1))
    if delta == 0:
        h = 0
    elif cmax == r_:
        h = 60 * (((g_ - b_) / delta) % 6)
    elif cmax == g_:
        h = 60 * ((b_ - r_) / delta + 2)
    else:
        h = 60 * ((r_ - g_) / delta + 4)
    return round(h, 2), round(s * 100, 2), round(l * 100, 2)


def handler(event):
    r = event.get("r") if isinstance(event, dict) else None
    g, b = event.get("g"), event.get("b")
    if r is None or g is None or b is None:
        return {"ok": False, "error": "r, g, b are required"}
    try:
        h, s, l = _rgb_to_hsl(int(r), int(g), int(b))
        return {"ok": True, "result": {"h": h, "s": s, "l": l}, "css": f"hsl({h},{s}%,{l}%)"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
