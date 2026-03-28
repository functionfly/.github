import math

def _mean(data): return sum(data) / len(data)
def _std(data, mean=None):
    if mean is None: mean = _mean(data)
    return math.sqrt(sum((x - mean) ** 2 for x in data) / len(data)) or 1e-9
def _median(data): s = sorted(data); n = len(s); return s[n//2] if n%2 else (s[n//2-1]+s[n//2])/2

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of numbers) is required"}
    try:
        nums = [float(x) for x in data if x is not None]
        if len(nums) < 4:
            return {"ok": False, "error": "data must contain at least 4 values"}
        method = event.get("method", "iqr")
        contamination = float(event.get("contamination", 0.1))
        outliers = []
        inliers = []
        if method == "iqr":
            sorted_nums = sorted(nums)
            n = len(sorted_nums)
            q1 = sorted_nums[n // 4]
            q3 = sorted_nums[3 * n // 4]
            iqr = q3 - q1
            lower = q1 - 1.5 * iqr
            upper = q3 + 1.5 * iqr
            for i, x in enumerate(nums):
                score = max(0, (lower - x) / (iqr or 1), (x - upper) / (iqr or 1))
                if x < lower or x > upper:
                    outliers.append({"index": i, "value": x, "score": round(score, 4), "lower_bound": round(lower, 4), "upper_bound": round(upper, 4)})
                else:
                    inliers.append(i)
        elif method == "zscore":
            mu = _mean(nums)
            sigma = _std(nums, mu)
            threshold = 3.0
            for i, x in enumerate(nums):
                z = abs((x - mu) / sigma)
                if z > threshold:
                    outliers.append({"index": i, "value": x, "score": round(z, 4), "zscore": round(z, 4)})
                else:
                    inliers.append(i)
        elif method == "isolation_forest":
            # Simplified: use contamination-based percentile
            sorted_with_idx = sorted(enumerate(nums), key=lambda x: abs(x[1] - _mean(nums)), reverse=True)
            n_outliers = max(1, int(len(nums) * contamination))
            outlier_indices = set(i for i, _ in sorted_with_idx[:n_outliers])
            for i, x in enumerate(nums):
                if i in outlier_indices:
                    outliers.append({"index": i, "value": x, "score": round(abs(x - _mean(nums)) / (_std(nums) or 1), 4)})
                else:
                    inliers.append(i)
        mu = _mean(nums)
        sigma = _std(nums, mu)
        return {
            "ok": True,
            "result": outliers,
            "outliers": outliers,
            "inlier_indices": inliers,
            "outlier_count": len(outliers),
            "outlier_ratio": round(len(outliers) / len(nums), 4),
            "method": method,
            "statistics": {"mean": round(mu, 4), "std": round(sigma, 4), "min": min(nums), "max": max(nums), "n": len(nums)}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
