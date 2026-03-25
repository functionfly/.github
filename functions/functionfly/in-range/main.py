def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    min_val = event.get("min")
    max_val = event.get("max")
    inclusive = event.get("inclusive", True)

    if value is None:
        return {"ok": False, "error": "value is required"}
    if min_val is None or max_val is None:
        return {"ok": False, "error": "min and max are required"}

    try:
        v = float(value)
        lo = float(min_val)
        hi = float(max_val)
    except (TypeError, ValueError):
        return {"ok": False, "error": "value, min, and max must be numbers"}

    if inclusive:
        result = lo <= v <= hi
    else:
        result = lo < v < hi

    return {"ok": True, "value": value, "result": result, "min": min_val, "max": max_val, "inclusive": inclusive}
