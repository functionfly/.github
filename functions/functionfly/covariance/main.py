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
        population = bool(event.get("population", False))
        mean_x = sum(x) / n
        mean_y = sum(y) / n
        cov = sum((x[i] - mean_x) * (y[i] - mean_y) for i in range(n))
        divisor = n if population else n - 1
        cov /= divisor
        return {"ok": True, "result": round(cov, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
