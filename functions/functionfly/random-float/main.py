import random

def handler(event):
    lo = event.get("min", 0) if isinstance(event, dict) else 0
    hi = event.get("max", 1) if isinstance(event, dict) else 1
    try:
        lo, hi = float(lo), float(hi)
        if lo > hi:
            lo, hi = hi, lo
        return {"ok": True, "value": random.uniform(lo, hi)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
