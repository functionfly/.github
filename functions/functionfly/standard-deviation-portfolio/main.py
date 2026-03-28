import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = [float(x) for x in returns]
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        population = bool(event.get("population", False))
        mean = sum(r) / n
        divisor = n if population else n - 1
        variance = sum((x - mean) ** 2 for x in r) / divisor
        std_dev = math.sqrt(variance)
        return {"ok": True, "result": round(std_dev, 8), "mean": round(mean, 8), "variance": round(variance, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
