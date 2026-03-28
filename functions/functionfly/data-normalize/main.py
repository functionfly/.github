def _flatten(data):
    """Flatten nested arrays."""
    if isinstance(data[0], list):
        return [x for row in data for x in row], True, len(data[0])
    return data, False, 1

def _normalize_1d(data, feature_range=(0, 1)):
    mn = min(data)
    mx = max(data)
    rng = mx - mn
    lo, hi = feature_range
    if rng == 0:
        return [lo] * len(data), mn, mx
    return [round(lo + (x - mn) / rng * (hi - lo), 6) for x in data], mn, mx

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array) is required"}
    try:
        feature_range = event.get("feature_range", [0, 1])
        lo, hi = float(feature_range[0]), float(feature_range[1])
        is_2d = isinstance(data[0], list)
        if is_2d:
            # Normalize each feature (column) independently
            n_features = len(data[0])
            normalized = [[0.0] * n_features for _ in range(len(data))]
            feature_stats = []
            for j in range(n_features):
                col = [float(data[i][j]) for i in range(len(data))]
                norm_col, mn, mx = _normalize_1d(col, (lo, hi))
                for i in range(len(data)):
                    normalized[i][j] = norm_col[i]
                feature_stats.append({"feature": j, "min": mn, "max": mx})
            return {
                "ok": True,
                "result": normalized,
                "normalized": normalized,
                "feature_range": [lo, hi],
                "feature_stats": feature_stats,
                "shape": [len(data), n_features]
            }
        else:
            nums = [float(x) for x in data if x is not None]
            normalized, mn, mx = _normalize_1d(nums, (lo, hi))
            return {
                "ok": True,
                "result": normalized,
                "normalized": normalized,
                "feature_range": [lo, hi],
                "original_min": mn,
                "original_max": mx,
                "n": len(normalized)
            }
    except Exception as e:
        return {"ok": False, "error": str(e)}
