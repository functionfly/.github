import hashlib
import itertools

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    param_grid = event.get("param_grid")
    if not param_grid or not isinstance(param_grid, dict):
        return {"ok": False, "error": "param_grid (object) is required"}
    try:
        metric = event.get("metric", "accuracy")
        cv = int(event.get("cv", 5))
        # Generate all combinations
        keys = list(param_grid.keys())
        values = [param_grid[k] if isinstance(param_grid[k], list) else [param_grid[k]] for k in keys]
        combinations = list(itertools.product(*values))
        if len(combinations) > 100:
            combinations = combinations[:100]
        results = []
        for combo in combinations:
            params = dict(zip(keys, combo))
            # Deterministic mock score based on params
            seed = str(sorted(params.items()))
            h = hashlib.sha256(seed.encode()).digest()
            score = round(0.5 + (h[0] / 255.0) * 0.5, 4)
            std = round(0.01 + (h[1] / 255.0) * 0.05, 4)
            results.append({
                "params": params,
                "mean_score": score,
                "std_score": std,
                "cv_scores": [round(score + (h[i % 32] / 255.0 - 0.5) * std * 2, 4) for i in range(cv)]
            })
        results.sort(key=lambda x: x["mean_score"], reverse=True)
        best = results[0]
        return {
            "ok": True,
            "result": best,
            "best_params": best["params"],
            "best_score": best["mean_score"],
            "metric": metric,
            "cv": cv,
            "n_combinations": len(results),
            "all_results": results[:20],
            "note": "Mock grid search — for production use, integrate scikit-learn GridSearchCV"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
