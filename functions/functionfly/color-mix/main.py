def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    colors = event.get("colors") if isinstance(event, dict) else None
    if not colors or not isinstance(colors, list) or len(colors) < 2:
        return {"ok": False, "error": "colors must be an array of at least 2 hex colors"}
    try:
        rs, gs, bs = [], [], []
        for c in colors:
            r,g,b = _parse(str(c)); rs.append(r); gs.append(g); bs.append(b)
        mr, mg, mb = round(sum(rs)/len(rs)), round(sum(gs)/len(gs)), round(sum(bs)/len(bs))
        return {"ok": True, "result": f"#{mr:02X}{mg:02X}{mb:02X}", "count": len(colors)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
