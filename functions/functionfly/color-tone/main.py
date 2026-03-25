def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    weight = float(event.get("weight", 0.5))
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r,g,b = _parse(str(color)); w = max(0.0, min(1.0, weight))
        grey = 128
        nr = round(r + (grey-r) * w); ng = round(g + (grey-g) * w); nb = round(b + (grey-b) * w)
        return {"ok": True, "result": f"#{nr:02X}{ng:02X}{nb:02X}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
