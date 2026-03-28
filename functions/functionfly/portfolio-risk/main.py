import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    weights = event.get("weights")
    std_devs = event.get("std_devs")
    corr_matrix = event.get("correlation_matrix")
    if weights is None or std_devs is None or corr_matrix is None:
        return {"ok": False, "error": "weights, std_devs, and correlation_matrix are required"}
    try:
        w = [float(x) for x in weights]
        s = [float(x) for x in std_devs]
        n = len(w)
        if len(s) != n or len(corr_matrix) != n:
            return {"ok": False, "error": "weights, std_devs, and correlation_matrix must have same length"}
        variance = 0.0
        for i in range(n):
            for j in range(n):
                variance += w[i] * w[j] * s[i] * s[j] * float(corr_matrix[i][j])
        std_dev = math.sqrt(variance)
        return {"ok": True, "result": round(std_dev, 8), "variance": round(variance, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
