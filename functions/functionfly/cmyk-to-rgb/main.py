def handler(event):
    c = event.get("c") if isinstance(event, dict) else None
    m, y, k = event.get("m"), event.get("y"), event.get("k")
    if any(v is None for v in [c, m, y, k]):
        return {"ok": False, "error": "c, m, y, k are required (0-100)"}
    try:
        c_, m_, y_, k_ = float(c)/100, float(m)/100, float(y)/100, float(k)/100
        r = round(255 * (1 - c_) * (1 - k_))
        g = round(255 * (1 - m_) * (1 - k_))
        b = round(255 * (1 - y_) * (1 - k_))
        return {"ok": True, "result": {"r": r, "g": g, "b": b}, "hex": f"#{r:02X}{g:02X}{b:02X}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
