import re


def handler(event):
    hex_color = event.get("hex") if isinstance(event, dict) else None
    include_alpha = event.get("include_alpha", True)

    if not hex_color:
        return {"ok": False, "error": "hex is required"}

    try:
        h = str(hex_color).strip().lstrip("#")
        if len(h) in (3, 4):
            h = "".join(c * 2 for c in h)
        if len(h) == 6:
            r, g, b = int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)
            a = 255
        elif len(h) == 8:
            r, g, b, a = int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), int(h[6:8], 16)
        else:
            return {"ok": False, "error": f"invalid hex color: #{h}"}

        result = {"r": r, "g": g, "b": b}
        if include_alpha and len(str(hex_color).lstrip("#")) in (4, 8):
            result["a"] = a
        result["hex"] = f"#{h.upper()}"
        result["css"] = f"rgb({r},{g},{b})" if a == 255 else f"rgba({r},{g},{b},{round(a/255,3)})"
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
