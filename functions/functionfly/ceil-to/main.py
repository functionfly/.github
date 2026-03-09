import math

def handler(event):
    if isinstance(event, dict):
        v = event.get("value")
        d = event.get("decimals", 0)
    else:
        v, d = None, 0
    if v is None:
        return {"ok": False, "error": "value is required"}
    try:
        v = float(v)
        d = int(d)
        m = 10 ** d
        return {"ok": True, "result": math.ceil(v * m) / m}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
