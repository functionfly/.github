def _parse_hex(h):
    h = h.strip().lstrip("#")
    if len(h) == 3: h = "".join(c*2 for c in h)
    return int(h[0:2],16), int(h[2:4],16), int(h[4:6],16)

def _luminance(r, g, b):
    def lin(c):
        c /= 255
        return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4
    return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    r = event.get("r")
    g = event.get("g")
    b = event.get("b")
    if not color and (r is None or g is None or b is None):
        return {"ok": False, "error": "provide 'color' (hex) or r, g, b"}
    try:
        if color:
            r, g, b = _parse_hex(str(color))
        lum = _luminance(int(r), int(g), int(b))
        return {"ok": True, "result": round(lum, 6), "luminance": round(lum, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
