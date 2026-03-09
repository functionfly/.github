def handler(event):
    arrays = event.get("arrays", [])
    method = event.get("method", "round_robin")  # "round_robin", "zip"

    if not isinstance(arrays, list):
        return {"ok": False, "error": "arrays must be an array of arrays"}

    if not arrays:
        return {"ok": True, "result": [], "count": 0}

    valid_methods = ["round_robin", "zip"]
    if method not in valid_methods:
        return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}

    # Validate that all elements are arrays
    for i, arr in enumerate(arrays):
        if not isinstance(arr, list):
            return {"ok": False, "error": f"arrays[{i}] must be an array"}

    try:
        result = []

        if method == "round_robin":
            # Round-robin interleaving
            max_len = max(len(arr) for arr in arrays) if arrays else 0

            for i in range(max_len):
                for arr in arrays:
                    if i < len(arr):
                        result.append(arr[i])

        elif method == "zip":
            # Zip-like interleaving (only up to shortest array)
            min_len = min(len(arr) for arr in arrays) if arrays else 0

            for i in range(min_len):
                for arr in arrays:
                    result.append(arr[i])

        return {
            "ok": True,
            "result": result,
            "count": len(result),
            "method": method,
            "array_count": len(arrays)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to interleave arrays: {str(e)}"}