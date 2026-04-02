def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    y_true = event.get("y_true")
    y_pred = event.get("y_pred")
    if not isinstance(y_true, list) or not isinstance(y_pred, list):
        return {"ok": False, "error": "y_true and y_pred must be arrays"}
    if len(y_true) != len(y_pred):
        return {"ok": False, "error": "y_true and y_pred must have same length"}
    if len(y_true) == 0:
        return {"ok": False, "error": "arrays must not be empty"}
    try:
        average = event.get("average", "macro")
        labels = sorted(set(str(l) for l in y_true + y_pred))
        n = len(y_true)
        # Overall accuracy
        correct = sum(1 for t, p in zip(y_true, y_pred) if t == p)
        accuracy = correct / n
        # Per-class metrics
        per_class = {}
        tp_total = fp_total = fn_total = 0
        weighted_precision_sum = 0
        weighted_recall_sum = 0
        weighted_f1_sum = 0
        support_total = 0
        for label in labels:
            tp = sum(
                1 for t, p in zip(y_true, y_pred) if str(t) == label and str(p) == label
            )
            fp = sum(
                1 for t, p in zip(y_true, y_pred) if str(t) != label and str(p) == label
            )
            fn = sum(
                1 for t, p in zip(y_true, y_pred) if str(t) == label and str(p) != label
            )
            support = sum(1 for t in y_true if str(t) == label)
            precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
            recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
            f1 = (
                2 * precision * recall / (precision + recall)
                if (precision + recall) > 0
                else 0.0
            )
            per_class[label] = {
                "precision": round(precision, 4),
                "recall": round(recall, 4),
                "f1": round(f1, 4),
                "support": support,
            }
            tp_total += tp
            fp_total += fp
            fn_total += fn
            weighted_precision_sum += precision * support
            weighted_recall_sum += recall * support
            weighted_f1_sum += f1 * support
            support_total += support
        # Macro averages
        n_classes = len(labels)
        macro_precision = sum(p["precision"] for p in per_class.values()) / n_classes
        macro_recall = sum(p["recall"] for p in per_class.values()) / n_classes
        macro_f1 = sum(p["f1"] for p in per_class.values()) / n_classes
        # Micro averages
        micro_precision = (
            tp_total / (tp_total + fp_total) if (tp_total + fp_total) > 0 else 0
        )
        micro_recall = (
            tp_total / (tp_total + fn_total) if (tp_total + fn_total) > 0 else 0
        )
        micro_f1 = (
            2 * micro_precision * micro_recall / (micro_precision + micro_recall)
            if (micro_precision + micro_recall) > 0
            else 0
        )
        # Weighted averages
        w_precision = weighted_precision_sum / support_total if support_total > 0 else 0
        w_recall = weighted_recall_sum / support_total if support_total > 0 else 0
        w_f1 = weighted_f1_sum / support_total if support_total > 0 else 0
        metrics = {
            "macro": {
                "precision": round(macro_precision, 4),
                "recall": round(macro_recall, 4),
                "f1": round(macro_f1, 4),
            },
            "micro": {
                "precision": round(micro_precision, 4),
                "recall": round(micro_recall, 4),
                "f1": round(micro_f1, 4),
            },
            "weighted": {
                "precision": round(w_precision, 4),
                "recall": round(w_recall, 4),
                "f1": round(w_f1, 4),
            },
        }
        return {
            "ok": True,
            "result": metrics,
            "accuracy": round(accuracy, 4),
            "per_class": per_class,
            "metrics": metrics,
            "average": average,
            "selected_metrics": metrics.get(average, metrics["macro"]),
            "n_samples": n,
            "n_classes": n_classes,
            "labels": labels,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
