import heapq


def handler(event):
    arrays = event.get("arrays", [])

    if not isinstance(arrays, list):
        return {"ok": False, "error": "arrays must be an array of arrays"}

    if not arrays:
        return {"ok": True, "result": [], "count": 0}

    # Validate that all elements are arrays
    for i, arr in enumerate(arrays):
        if not isinstance(arr, list):
            return {"ok": False, "error": f"arrays[{i}] must be an array"}

    try:
        # Use heapq.merge for efficient merging of multiple sorted arrays
        merged = list(heapq.merge(*arrays))

        return {
            "ok": True,
            "result": merged,
            "count": len(merged),
            "array_count": len(arrays),
            "total_input_elements": sum(len(arr) for arr in arrays)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to merge sorted arrays: {str(e)}"}