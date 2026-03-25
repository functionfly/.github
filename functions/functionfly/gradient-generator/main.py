def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    color1 = event.get("color1") if isinstance(event, dict) else None
    color2 = event.get("color2")
    steps = int(event.get("steps", 5))
    direction = event.get("direction", "to right")
    if not color1 or not color2:
        return {"ok": False, "error": "color1 and color2 are required"}
    if steps < 2 or steps > 50:
        return {"ok": False, "error": "steps must be between 2 and 50"}
    try:
        r1,g1,b1=_parse(str(color1)); r2,g2,b2=_parse(str(color2))
        stops = []
        for i in range(steps):
            t = i / (steps-1)
            r=round(r1+(r2-r1)*t); g=round(g1+(g2-g1)*t); b=round(b1+(b2-b1)*t)
            pct = round(t*100)
            stops.append({"hex": f"#{r:02X}{g:02X}{b:02X}", "percent": pct})
        css = f"linear-gradient({direction}, {', '.join(s['hex'] for s in stops)})"
        return {"ok": True, "result": css, "stops": stops}
    except Exception as e:
        return {"ok": False, "error": str(e)}
