import math

def _standardize_1d(data):
    n = len(data)
    mean = sum(data) / n
    std = math.sqrt(sum((x - mean) ** 2 for x in data) / n) or 1e-9
    return [round((x - mean) / std, 6) for x in data], mean, std

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array) is required"}
    try:
        is_2d = isinstance(data[0], list)
        if is_2d:
            n_features = len(data[0])
            standardized = [[0.0] * n_features for _ in range(len(data))]
            feature_stats = []
            for j in range(n_features):
                col = [float(data[i][j]) for i in range(len(data))]
                std_col, mean, std = _standardize_1d(col)
                for i in range(len(data)):
                    standardized[i][j] = std_col[i]
                feature_stats.append({"feature": j, "mean": round(mean, 6), "std": round(std, 6)})
            return {
                "ok": True,
                "result": standardized,
                "standardized": standardized,
                "feature_stats": feature_stats,
                "shape": [len(data), n_features]
            }
        else:
            nums = [float(x) for x in data if x is not None]
            standardized, mean, std = _standardize_1d(nums)
            return {
                "ok": True,
                "result": standardized,
                "standardized": standardized,
                "mean": round(mean, 6),
                "std": round(std, 6),
                "n": len(standardized)
            }
    except Exception as e:
        return {"ok": False, "error": str(e)}
