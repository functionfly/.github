def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    y_true = event.get("y_true")
    y_score = event.get("y_score")
    if not isinstance(y_true, list) or not isinstance(y_score, list):
        return {"ok": False, "error": "y_true and y_score must be arrays"}
    if len(y_true) != len(y_score):
        return {"ok": False, "error": "y_true and y_score must have same length"}
    if len(y_true) == 0:
        return {"ok": False, "error": "arrays must not be empty"}
    try:
        n_pos = sum(1 for y in y_true if y == 1)
        n_neg = sum(1 for y in y_true if y == 0)
        if n_pos == 0 or n_neg == 0:
            return {"ok": False, "error": "both classes (0 and 1) must be present"}
        # Sort by score descending
        pairs = sorted(zip(y_score, y_true), key=lambda x: -x[0])
        # Generate thresholds
        thresholds = sorted(set(y_score), reverse=True)
        thresholds = [float("inf")] + thresholds + [-float("inf")]
        num_thresh = int(event.get("num_thresholds", 100))
        if len(thresholds) > num_thresh + 2:
            step = max(1, len(thresholds) // num_thresh)
            thresholds = thresholds[::step]
            if thresholds[-1] != -float("inf"):
                thresholds.append(-float("inf"))
        roc_points = []
        for thresh in thresholds:
            tp = sum(1 for s, y in pairs if s >= thresh and y == 1)
            fp = sum(1 for s, y in pairs if s >= thresh and y == 0)
            tpr = tp / n_pos
            fpr = fp / n_neg
            roc_points.append(
                {
                    "fpr": round(fpr, 6),
                    "tpr": round(tpr, 6),
                    "threshold": thresh if thresh != float("inf") else None,
                }
            )
        # AUC using trapezoidal rule (sorted by FPR ascending)
        roc_sorted = sorted(roc_points, key=lambda p: p["fpr"])
        auc = 0.0
        for i in range(1, len(roc_sorted)):
            dx = roc_sorted[i]["fpr"] - roc_sorted[i - 1]["fpr"]
            y_avg = (roc_sorted[i]["tpr"] + roc_sorted[i - 1]["tpr"]) / 2
            auc += dx * y_avg
        # Find optimal threshold (Youden's J = TPR - FPR)
        best_j = -1
        best_thresh = None
        best_point = None
        for p in roc_points:
            if p["threshold"] is not None:
                j = p["tpr"] - p["fpr"]
                if j > best_j:
                    best_j = j
                    best_thresh = p["threshold"]
                    best_point = p
        return {
            "ok": True,
            "result": roc_points,
            "auc": round(auc, 4),
            "roc_points": roc_points,
            "optimal_threshold": best_thresh,
            "optimal_point": best_point,
            "youdens_j": round(best_j, 4),
            "n_positives": n_pos,
            "n_negatives": n_neg,
            "n_samples": len(y_true),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
