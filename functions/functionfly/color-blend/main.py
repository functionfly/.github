def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    color1 = event.get("color1") if isinstance(event, dict) else None
    color2 = event.get("color2")
    weight = float(event.get("weight", 0.5))
    if not color1 or not color2:
        return {"ok": False, "error": "color1 and color2 are required"}
    try:
        r1,g1,b1 = _parse(str(color1)); r2,g2,b2 = _parse(str(color2))
        w = max(0.0, min(1.0, weight))
        r = round(r1 * w + r2 * (1-w))
        g = round(g1 * w + g2 * (1-w))
        b = round(b1 * w + b2 * (1-w))
        return {"ok": True, "result": f"#{r:02X}{g:02X}{b:02X}", "r": r, "g": g, "b": b}
    except Exception as e:
        return {"ok": False, "error": str(e)}
