def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    threshold = float(event.get("threshold", 0.5))
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r,g,b = _parse(str(color))
        def lin(c):
            c/=255; return c/12.92 if c<=0.03928 else ((c+0.055)/1.055)**2.4
        lum = 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
        return {"ok": True, "result": lum < threshold, "luminance": round(lum,4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
