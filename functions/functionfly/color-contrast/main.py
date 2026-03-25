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
    color1 = event.get("color1") if isinstance(event, dict) else None
    color2 = event.get("color2")
    if not color1 or not color2:
        return {"ok": False, "error": "color1 and color2 are required (hex)"}
    try:
        r1,g1,b1 = _parse_hex(str(color1))
        r2,g2,b2 = _parse_hex(str(color2))
        l1, l2 = _luminance(r1,g1,b1), _luminance(r2,g2,b2)
        lighter, darker = max(l1,l2), min(l1,l2)
        ratio = round((lighter + 0.05) / (darker + 0.05), 2)
        return {
            "ok": True,
            "result": ratio,
            "ratio": f"{ratio}:1",
            "passes_aa": ratio >= 4.5,
            "passes_aa_large": ratio >= 3.0,
            "passes_aaa": ratio >= 7.0,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
