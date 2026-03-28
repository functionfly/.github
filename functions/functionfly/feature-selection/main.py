import math

def _variance(col):
    n = len(col)
    mean = sum(col) / n
    return sum((x - mean) ** 2 for x in col) / n

def _correlation(x, y):
    n = len(x)
    mx = sum(x) / n
    my = sum(y) / n
    num = sum((xi - mx) * (yi - my) for xi, yi in zip(x, y))
    dx = math.sqrt(sum((xi - mx) ** 2 for xi in x)) or 1e-9
    dy = math.sqrt(sum((yi - my) ** 2 for yi in y)) or 1e-9
    return num / (dx * dy)

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of arrays) is required"}
    try:
        labels = event.get("labels")
        method = event.get("method", "variance")
        k = int(event.get("k", 5))
        processed = [[float(x) for x in row] for row in data if isinstance(row, list)]
        if len(processed) < 2:
            return {"ok": False, "error": "data must contain at least 2 samples"}
        n_features = len(processed[0])
        k = min(k, n_features)
        feature_scores = []
        for j in range(n_features):
            col = [processed[i][j] for i in range(len(processed))]
            if method == "variance":
                score = _variance(col)
            elif method == "correlation" and labels:
                label_nums = [float(l) for l in labels]
                score = abs(_correlation(col, label_nums))
            elif method == "range":
                score = max(col) - min(col)
            else:
                score = _variance(col)
            feature_scores.append({"feature_index": j, "score": round(score, 6), "method": method})
        feature_scores.sort(key=lambda x: x["score"], reverse=True)
        selected = feature_scores[:k]
        selected_indices = [f["feature_index"] for f in selected]
        # Filter data to selected features
        filtered_data = [[row[j] for j in selected_indices] for row in processed]
        return {
            "ok": True,
            "result": {"selected_indices": selected_indices, "scores": selected},
            "selected_indices": selected_indices,
            "feature_scores": feature_scores,
            "selected_features": selected,
            "filtered_data": filtered_data,
            "k": k,
            "method": method
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
