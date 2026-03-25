def _parse(h):
    h = h.strip().lstrip("#")
    if len(h) == 3: h = "".join(c*2 for c in h)
    return int(h[0:2],16), int(h[2:4],16), int(h[4:6],16)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r, g, b = _parse(str(color))
        ir, ig, ib = 255 - r, 255 - g, 255 - b
        return {"ok": True, "result": f"#{ir:02X}{ig:02X}{ib:02X}", "r": ir, "g": ig, "b": ib}
    except Exception as e:
        return {"ok": False, "error": str(e)}
