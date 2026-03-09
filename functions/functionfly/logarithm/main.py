import math

def handler(event):
    v = event.get("value") if isinstance(event, dict) else None
    base = event.get("base") if isinstance(event, dict) else None
    if v is None:
        return {"ok": False, "error": "value is required"}
    try:
        v = float(v)
        if v <= 0:
            return {"ok": False, "error": "value must be positive"}
        if base is None:
            return {"ok": True, "result": math.log(v)}
        base = float(base)
        if base <= 0 or base == 1:
            return {"ok": False, "error": "base must be positive and not 1"}
        return {"ok": True, "result": math.log(v, base)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
