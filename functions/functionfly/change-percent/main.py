def handler(event):
    if isinstance(event, dict):
        old = event.get("old")
        new = event.get("new")
    else:
        old, new = None, None
    if old is None or new is None:
        return {"ok": False, "error": "old and new are required"}
    try:
        old, new = float(old), float(new)
        if old == 0:
            return {"ok": False, "error": "old cannot be zero"}
        return {"ok": True, "change_percent": round((new - old) / old * 100, 10)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
