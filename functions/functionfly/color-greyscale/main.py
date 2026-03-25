def _parse(h):
    h = h.strip().lstrip("#")
    if len(h) == 3: h = "".join(c*2 for c in h)
    return int(h[0:2],16), int(h[2:4],16), int(h[4:6],16)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    method = event.get("method", "luma")
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r, g, b = _parse(str(color))
        if method == "average":
            v = round((r + g + b) / 3)
        elif method == "lightness":
            v = round((max(r,g,b) + min(r,g,b)) / 2)
        else:
            v = round(0.2126 * r + 0.7152 * g + 0.0722 * b)
        return {"ok": True, "result": {"r": v, "g": v, "b": v}, "hex": f"#{v:02X}{v:02X}{v:02X}", "value": v}
    except Exception as e:
        return {"ok": False, "error": str(e)}
