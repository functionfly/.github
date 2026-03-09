import math

def handler(event):
    d = event.get("degrees") if isinstance(event, dict) else None
    if d is None:
        return {"ok": False, "error": "degrees is required"}
    try:
        return {"ok": True, "radians": math.radians(float(d))}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
