def handler(event):
    if isinstance(event, dict):
        part = event.get("part")
        whole = event.get("whole")
    else:
        part, whole = None, None
    if part is None or whole is None:
        return {"ok": False, "error": "part and whole are required"}
    try:
        part, whole = float(part), float(whole)
        if whole == 0:
            return {"ok": False, "error": "whole cannot be zero"}
        return {"ok": True, "percentage": round(part / whole * 100, 10)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
