def handler(event):
    r = event.get("r") if isinstance(event, dict) else None
    g = event.get("g")
    b = event.get("b")
    a = event.get("a")
    uppercase = event.get("uppercase", True)

    if r is None or g is None or b is None:
        return {"ok": False, "error": "r, g, b are required"}

    try:
        ri, gi, bi = int(r), int(g), int(b)
        for name, v in [("r", ri), ("g", gi), ("b", bi)]:
            if not 0 <= v <= 255:
                return {"ok": False, "error": f"{name} must be 0-255"}

        if a is not None:
            ai = int(a)
            if not 0 <= ai <= 255:
                return {"ok": False, "error": "a must be 0-255"}
            hex_str = f"#{ri:02x}{gi:02x}{bi:02x}{ai:02x}"
        else:
            hex_str = f"#{ri:02x}{gi:02x}{bi:02x}"

        if uppercase:
            hex_str = hex_str.upper()

        return {"ok": True, "result": hex_str, "r": ri, "g": gi, "b": bi}
    except Exception as e:
        return {"ok": False, "error": str(e)}
