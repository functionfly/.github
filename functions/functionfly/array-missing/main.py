def handler(event):
    items = event.get("items", [])
    method = event.get("method", "count")  # "count", "indices", "values", "mask"

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    valid_methods = ["count", "indices", "values", "mask"]
    if method not in valid_methods:
        return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}

    try:
        missing_count = 0
        missing_indices = []
        missing_values = []
        mask = []

        for i, item in enumerate(items):
            is_missing = item is None or (isinstance(item, float) and str(item).lower() in ['nan', 'inf', '-inf'])

            if is_missing:
                missing_count += 1
                missing_indices.append(i)
                missing_values.append(item)
                mask.append(True)
            else:
                mask.append(False)

        if method == "count":
            result = missing_count
        elif method == "indices":
            result = missing_indices
        elif method == "values":
            result = missing_values
        elif method == "mask":
            result = mask

        return {
            "ok": True,
            "result": result,
            "method": method,
            "total_count": len(items),
            "missing_count": missing_count,
            "missing_percentage": (missing_count / len(items) * 100) if items else 0
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to analyze missing values: {str(e)}"}