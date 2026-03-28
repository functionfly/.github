import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    if x is None:
        return {"ok": False, "error": "x is required"}
    try:
        x = float(x)
        mu = float(event.get("mu", 0))
        sigma = float(event.get("sigma", 1))
        if x <= 0:
            return {"ok": False, "error": "x must be positive"}
        if sigma <= 0:
            return {"ok": False, "error": "sigma must be positive"}
        pdf = (math.exp(-(math.log(x) - mu) ** 2 / (2 * sigma ** 2))
               / (x * sigma * math.sqrt(2 * math.pi)))
        z = (math.log(x) - mu) / sigma
        cdf = 0.5 * (1 + math.erf(z / math.sqrt(2)))
        return {"ok": True, "result": {"pdf": round(pdf, 8), "cdf": round(cdf, 8)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
