import math

def handler(event):
    r = event.get("radians") if isinstance(event, dict) else None
    if r is None:
        return {"ok": False, "error": "radians is required"}
    try:
        return {"ok": True, "degrees": math.degrees(float(r))}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
