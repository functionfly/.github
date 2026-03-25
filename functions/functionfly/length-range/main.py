def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    min_len = event.get("min", 0)
    max_len = event.get("max")

    if value is None and "value" not in event:
        return {"ok": False, "error": "value is required"}
    if max_len is None:
        return {"ok": False, "error": "max is required"}

    if isinstance(value, (str, list, dict, tuple)):
        length = len(value)
    else:
        length = len(str(value))

    try:
        lo = int(min_len)
        hi = int(max_len)
    except (TypeError, ValueError):
        return {"ok": False, "error": "min and max must be integers"}

    result = lo <= length <= hi
    return {"ok": True, "value": value, "result": result, "length": length, "min": lo, "max": hi}
