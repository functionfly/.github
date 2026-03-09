import math

def handler(event):
    v = event.get("value") if isinstance(event, dict) else None
    if v is None:
        return {"ok": False, "error": "value is required"}
    try:
        v = float(v)
        if v < 0:
            return {"ok": False, "error": "value must be non-negative"}
        return {"ok": True, "result": math.sqrt(v)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
