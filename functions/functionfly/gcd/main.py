import math

def handler(event):
    a = event.get("a") if isinstance(event, dict) else None
    b = event.get("b") if isinstance(event, dict) else None
    if a is None or b is None:
        return {"ok": False, "error": "a and b are required"}
    try:
        a, b = int(a), int(b)
        return {"ok": True, "result": math.gcd(a, b)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
