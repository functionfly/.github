import itertools


def handler(event):
    items = event.get("items", [])
    r = event.get("r")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if r is None:
        return {"ok": False, "error": "r (combination length) is required"}

    if not isinstance(r, int) or r < 0:
        return {"ok": False, "error": "r must be a non-negative integer"}

    if r > len(items):
        return {"ok": False, "error": "r cannot be greater than the length of items"}

    try:
        combinations = list(itertools.combinations(items, r))

        # Convert tuples back to lists
        result = [list(c) for c in combinations]

        return {
            "ok": True,
            "result": result,
            "count": len(result)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to generate combinations: {str(e)}"}
