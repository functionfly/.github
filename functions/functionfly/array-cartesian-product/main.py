import itertools


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
        cartesian_product = list(itertools.product(*arrays))

        # Convert tuples back to lists
        result = [list(cp) for cp in cartesian_product]

        return {
            "ok": True,
            "result": result,
            "count": len(result)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to generate cartesian product: {str(e)}"}
