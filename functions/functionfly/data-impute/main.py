import math

def _median(values):
    s = sorted(values)
    n = len(s)
    return s[n // 2] if n % 2 else (s[n // 2 - 1] + s[n // 2]) / 2

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array) is required"}
    try:
        strategy = event.get("strategy", "mean")
        fill_value = event.get("fill_value")
        is_2d = isinstance(data[0], list) if data else False
        if is_2d:
            n_features = len(data[0])
            imputed = [list(row) for row in data]
            stats = []
            for j in range(n_features):
                col = [float(data[i][j]) for i in range(len(data)) if data[i][j] is not None]
                if not col:
                    fill = 0.0
                elif strategy == "mean":
                    fill = sum(col) / len(col)
                elif strategy == "median":
                    fill = _median(col)
                elif strategy == "mode":
                    from collections import Counter
                    fill = Counter(col).most_common(1)[0][0]
                elif strategy == "constant":
                    fill = float(fill_value) if fill_value is not None else 0.0
                elif strategy == "min":
                    fill = min(col)
                elif strategy == "max":
                    fill = max(col)
                else:
                    fill = sum(col) / len(col)
                missing_count = sum(1 for i in range(len(data)) if data[i][j] is None)
                for i in range(len(data)):
                    if data[i][j] is None:
                        imputed[i][j] = round(fill, 6)
                stats.append({"feature": j, "fill_value": round(fill, 6), "missing_count": missing_count})
            return {"ok": True, "result": imputed, "imputed": imputed, "strategy": strategy, "feature_stats": stats}
        else:
            nums = [x for x in data]
            valid = [float(x) for x in nums if x is not None]
            if not valid:
                return {"ok": False, "error": "No valid values found"}
            if strategy == "mean":
                fill = sum(valid) / len(valid)
            elif strategy == "median":
                fill = _median(valid)
            elif strategy == "constant":
                fill = float(fill_value) if fill_value is not None else 0.0
            elif strategy == "min":
                fill = min(valid)
            elif strategy == "max":
                fill = max(valid)
            else:
                fill = sum(valid) / len(valid)
            missing_count = sum(1 for x in nums if x is None)
            imputed = [round(fill, 6) if x is None else float(x) for x in nums]
            return {
                "ok": True,
                "result": imputed,
                "imputed": imputed,
                "strategy": strategy,
                "fill_value": round(fill, 6),
                "missing_count": missing_count,
                "missing_ratio": round(missing_count / len(nums), 4)
            }
    except Exception as e:
        return {"ok": False, "error": str(e)}
