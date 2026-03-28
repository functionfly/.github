import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    y = event.get("y")
    if x is None or y is None:
        return {"ok": False, "error": "x and y are required"}
    try:
        x = [float(v) for v in x]
        y = [float(v) for v in y]
        n = len(x)
        if n != len(y):
            return {"ok": False, "error": "x and y must have same length"}
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        mean_x = sum(x) / n
        mean_y = sum(y) / n
        cov = sum((x[i] - mean_x) * (y[i] - mean_y) for i in range(n))
        std_x = math.sqrt(sum((v - mean_x) ** 2 for v in x))
        std_y = math.sqrt(sum((v - mean_y) ** 2 for v in y))
        if std_x == 0 or std_y == 0:
            return {"ok": False, "error": "Standard deviation cannot be zero"}
        corr = cov / (std_x * std_y)
        return {"ok": True, "result": round(corr, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
