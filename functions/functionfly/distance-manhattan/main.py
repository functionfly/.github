def handler(event):
    if isinstance(event, dict):
        x1 = event.get("x1")
        y1 = event.get("y1")
        x2 = event.get("x2")
        y2 = event.get("y2")
    else:
        x1 = y1 = x2 = y2 = None
    if x1 is None or y1 is None or x2 is None or y2 is None:
        return {"ok": False, "error": "x1, y1, x2, y2 are required"}
    try:
        x1, y1 = float(x1), float(y1)
        x2, y2 = float(x2), float(y2)
        return {"ok": True, "distance": abs(x2 - x1) + abs(y2 - y1)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
