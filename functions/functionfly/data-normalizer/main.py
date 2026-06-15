import statistics
import math
from typing import Any


def calculate_stats(data: list) -> dict:
    clean_data = [x for x in data if isinstance(x, (int, float)) and x is not None and not math.isnan(x) and not math.isinf(x)]
    
    if len(clean_data) == 0:
        return {"mean": 0, "std": 0, "min": 0, "max": 0, "median": 0, "count": 0}
    
    mean = statistics.mean(clean_data)
    std = statistics.stdev(clean_data) if len(clean_data) > 1 else 0
    median = statistics.median(clean_data)
    min_val = min(clean_data)
    max_val = max(clean_data)
    
    return {
        "mean": round(mean, 6),
        "std": round(std, 6),
        "min": round(min_val, 6),
        "max": round(max_val, 6),
        "median": round(median, 6),
        "count": len(clean_data)
    }


def normalize_minmax(data: list) -> tuple[list, dict]:
    clean_data = [x for x in data if isinstance(x, (int, float)) and x is not None]
    
    if len(clean_data) == 0:
        return [], {"min": 0, "max": 0}
    
    min_val = min(clean_data)
    max_val = max(clean_data)
    range_val = max_val - min_val
    
    if range_val == 0:
        return [0.5] * len(data), {"min": min_val, "max": max_val, "range": 0}
    
    normalized = []
    for x in data:
        if isinstance(x, (int, float)) and x is not None:
            normalized.append(round((x - min_val) / range_val, 6))
        else:
            normalized.append(None)
    
    return normalized, {"min": min_val, "max": max_val, "range": range_val}


def normalize_zscore(data: list) -> tuple[list, dict]:
    clean_data = [x for x in data if isinstance(x, (int, float)) and x is not None]
    
    if len(clean_data) < 2:
        if len(clean_data) == 1:
            return [0.0] * len(data), {"mean": clean_data[0], "std": 0, "count": 1}
        return [None] * len(data), {"mean": 0, "std": 0, "count": 0}
    
    mean = statistics.mean(clean_data)
    std = statistics.stdev(clean_data)
    
    if std == 0:
        return [0.0] * len(data), {"mean": mean, "std": 0, "count": len(clean_data)}
    
    normalized = []
    for x in data:
        if isinstance(x, (int, float)) and x is not None:
            normalized.append(round((x - mean) / std, 6))
        else:
            normalized.append(None)
    
    return normalized, {"mean": round(mean, 6), "std": round(std, 6), "count": len(clean_data)}


def normalize_robust(data: list) -> tuple[list, dict]:
    clean_data = [x for x in data if isinstance(x, (int, float)) and x is not None]
    
    if len(clean_data) == 0:
        return [], {"median": 0, "iqr": 0}
    
    median = statistics.median(clean_data)
    
    q1 = statistics.quantiles(clean_data, n=4)[0] if len(clean_data) >= 4 else min(clean_data)
    q3 = statistics.quantiles(clean_data, n=4)[2] if len(clean_data) >= 4 else max(clean_data)
    iqr = q3 - q1
    
    if iqr == 0:
        iqr = max(clean_data) - min(clean_data)
        if iqr == 0:
            return [0.0] * len(data), {"median": median, "iqr": 0, "q1": q1, "q3": q3}
    
    normalized = []
    for x in data:
        if isinstance(x, (int, float)) and x is not None:
            normalized.append(round((x - median) / iqr, 6))
        else:
            normalized.append(None)
    
    return normalized, {
        "median": round(median, 6),
        "iqr": round(iqr, 6),
        "q1": round(q1, 6),
        "q3": round(q3, 6)
    }


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        data = event.get("data", [])
        method = event.get("method", "minmax").lower().strip()
        
        if not isinstance(data, list):
            return {"ok": False, "error": "data must be a list"}
        
        if len(data) == 0:
            return {"ok": False, "error": "data cannot be empty"}
        
        valid_methods = ["minmax", "z-score", "zscore", "robust"]
        if method not in valid_methods:
            return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}
        
        original_stats = calculate_stats(data)
        
        if method == "minmax":
            normalized_data, normalization_params = normalize_minmax(data)
        elif method in ["z-score", "zscore"]:
            normalized_data, normalization_params = normalize_zscore(data)
        else:
            normalized_data, normalization_params = normalize_robust(data)
        
        return {
            "ok": True,
            "normalized_data": normalized_data,
            "original_stats": original_stats,
            "normalization_params": normalization_params,
            "method": method
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
