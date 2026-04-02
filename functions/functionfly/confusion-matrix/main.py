def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    y_true = event.get("y_true")
    y_pred = event.get("y_pred")
    if not isinstance(y_true, list) or not isinstance(y_pred, list):
        return {"ok": False, "error": "y_true and y_pred must be arrays"}
    if len(y_true) != len(y_pred):
        return {"ok": False, "error": "y_true and y_pred must have the same length"}
    if len(y_true) == 0:
        return {"ok": False, "error": "arrays must not be empty"}
    try:
        labels = event.get("labels") or sorted(set(str(l) for l in y_true + y_pred))
        n_labels = len(labels)
        label_to_idx = {l: i for i, l in enumerate(labels)}
        # Build confusion matrix
        matrix = [[0] * n_labels for _ in range(n_labels)]
        for t, p in zip(y_true, y_pred):
            t_str = str(t)
            p_str = str(p)
            if t_str in label_to_idx and p_str in label_to_idx:
                matrix[label_to_idx[t_str]][label_to_idx[p_str]] += 1
        # Per-class metrics
        per_class = {}
        for i, label in enumerate(labels):
            tp = matrix[i][i]
            fp = sum(matrix[j][i] for j in range(n_labels)) - tp
            fn = sum(matrix[i]) - tp
            precision = tp / (tp + fp) if (tp + fp) > 0 else 0
            recall = tp / (tp + fn) if (tp + fn) > 0 else 0
            f1 = (
                2 * precision * recall / (precision + recall)
                if (precision + recall) > 0
                else 0
            )
            per_class[label] = {
                "tp": tp,
                "fp": fp,
                "fn": fn,
                "precision": round(precision, 4),
                "recall": round(recall, 4),
                "f1": round(f1, 4),
            }
        total = len(y_true)
        correct = sum(1 for t, p in zip(y_true, y_pred) if t == p)
        accuracy = correct / total if total > 0 else 0
        return {
            "ok": True,
            "result": matrix,
            "confusion_matrix": matrix,
            "labels": labels,
            "per_class": per_class,
            "accuracy": round(accuracy, 4),
            "total_samples": total,
            "correct": correct,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
