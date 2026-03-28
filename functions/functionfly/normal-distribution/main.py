import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    if x is None:
        return {"ok": False, "error": "x is required"}
    try:
        x = float(x)
        mu = float(event.get("mean", 0))
        sigma = float(event.get("std_dev", 1))
        if sigma <= 0:
            return {"ok": False, "error": "std_dev must be positive"}
        z = (x - mu) / sigma
        pdf = math.exp(-0.5 * z ** 2) / (sigma * math.sqrt(2 * math.pi))
        cdf = 0.5 * (1 + math.erf(z / math.sqrt(2)))
        return {"ok": True, "result": {"pdf": round(pdf, 8), "cdf": round(cdf, 8)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
