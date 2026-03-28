import math

def _mean(data):
    return sum(data) / len(data)

def _std(data, mean=None):
    if mean is None:
        mean = _mean(data)
    variance = sum((x - mean) ** 2 for x in data) / len(data)
    return math.sqrt(variance)

def _iqr_bounds(data):
    sorted_data = sorted(data)
    n = len(sorted_data)
    q1 = sorted_data[n // 4]
    q3 = sorted_data[3 * n // 4]
    iqr = q3 - q1
    return q1 - 1.5 * iqr, q3 + 1.5 * iqr

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of numbers) is required"}
    try:
        nums = [float(x) for x in data if x is not None]
        if len(nums) < 2:
            return {"ok": False, "error": "data must contain at least 2 values"}
        method = event.get("method", "zscore")
        threshold = float(event.get("threshold", 3.0))
        anomalies = []
        if method == "zscore":
            mu = _mean(nums)
            sigma = _std(nums, mu) or 1e-9
            for i, x in enumerate(nums):
                z = abs((x - mu) / sigma)
                if z > threshold:
                    anomalies.append({"index": i, "value": x, "score": round(z, 4), "method": "zscore"})
        elif method == "iqr":
            lower, upper = _iqr_bounds(nums)
            for i, x in enumerate(nums):
                if x < lower or x > upper:
                    score = max(abs(x - lower), abs(x - upper))
                    anomalies.append({"index": i, "value": x, "score": round(score, 4), "method": "iqr"})
        elif method == "mad":
            median = sorted(nums)[len(nums) // 2]
            mad = sorted([abs(x - median) for x in nums])[len(nums) // 2] or 1e-9
            for i, x in enumerate(nums):
                score = abs(x - median) / mad
                if score > threshold:
                    anomalies.append({"index": i, "value": x, "score": round(score, 4), "method": "mad"})
        mu = _mean(nums)
        sigma = _std(nums, mu)
        return {
            "ok": True,
            "result": anomalies,
            "anomalies": anomalies,
            "anomaly_count": len(anomalies),
            "anomaly_ratio": round(len(anomalies) / len(nums), 4),
            "method": method,
            "threshold": threshold,
            "statistics": {"mean": round(mu, 4), "std": round(sigma, 4), "min": min(nums), "max": max(nums), "n": len(nums)}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
