import hashlib
import math

def _accuracy(y_true, y_pred):
    return sum(1 for a, b in zip(y_true, y_pred) if a == b) / len(y_true)

def _mean(vals): return sum(vals) / len(vals)
def _std(vals, mean=None):
    if mean is None: mean = _mean(vals)
    return math.sqrt(sum((x - mean) ** 2 for x in vals) / len(vals))

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    labels = event.get("labels")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array) is required"}
    if not labels or not isinstance(labels, list):
        return {"ok": False, "error": "labels (array) is required"}
    if len(data) != len(labels):
        return {"ok": False, "error": "data and labels must have the same length"}
    try:
        k = int(event.get("k", 5))
        metric = event.get("metric", "accuracy")
        n = len(data)
        k = min(k, n)
        fold_size = n // k
        fold_scores = []
        for fold in range(k):
            # Create train/test split
            test_start = fold * fold_size
            test_end = test_start + fold_size if fold < k - 1 else n
            test_indices = list(range(test_start, test_end))
            train_indices = [i for i in range(n) if i not in set(test_indices)]
            # Mock prediction: use hash-based deterministic "model"
            test_labels = [labels[i] for i in test_indices]
            pred_labels = []
            for i in test_indices:
                seed = str(data[i]) + str(fold)
                h = hashlib.sha256(seed.encode()).digest()
                # Predict based on majority class with some noise
                unique_labels = list(set(labels))
                pred = unique_labels[h[0] % len(unique_labels)]
                pred_labels.append(pred)
            if metric == "accuracy":
                score = _accuracy(test_labels, pred_labels)
            else:
                score = _accuracy(test_labels, pred_labels)
            fold_scores.append(round(score, 4))
        mean_score = round(_mean(fold_scores), 4)
        std_score = round(_std(fold_scores, mean_score), 4)
        return {
            "ok": True,
            "result": {"mean_score": mean_score, "std_score": std_score, "fold_scores": fold_scores},
            "mean_score": mean_score,
            "std_score": std_score,
            "fold_scores": fold_scores,
            "k": k,
            "metric": metric,
            "n_samples": n,
            "confidence_interval": [round(mean_score - 2 * std_score, 4), round(mean_score + 2 * std_score, 4)],
            "note": "Mock cross-validation — for production use, integrate scikit-learn cross_val_score"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
