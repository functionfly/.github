import random

def handler(event):
    lo = event.get("min", 0) if isinstance(event, dict) else 0
    hi = event.get("max", 100) if isinstance(event, dict) else 100
    try:
        lo, hi = int(lo), int(hi)
        if lo > hi:
            lo, hi = hi, lo
        return {"ok": True, "value": random.randint(lo, hi)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
