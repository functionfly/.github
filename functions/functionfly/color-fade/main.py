def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    amount = float(event.get("amount", 0.5))
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r,g,b=_parse(str(color))
        alpha = max(0.0, min(1.0, 1.0 - amount))
        return {"ok": True, "result": f"rgba({r},{g},{b},{round(alpha,3)})", "alpha": round(alpha,3), "r": r, "g": g, "b": b}
    except Exception as e:
        return {"ok": False, "error": str(e)}
