import itertools


def handler(event):
    items = event.get("items", [])
    r = event.get("r")  # Length of permutations (optional)

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if r is not None:
        if not isinstance(r, int) or r < 0:
            return {"ok": False, "error": "r must be a non-negative integer"}
        if r > len(items):
            return {"ok": False, "error": "r cannot be greater than the length of items"}

    try:
        if r is not None:
            # Generate permutations of length r
            permutations = list(itertools.permutations(items, r))
        else:
            # Generate all permutations
            permutations = list(itertools.permutations(items))

        # Convert tuples back to lists
        result = [list(p) for p in permutations]

        return {
            "ok": True,
            "result": result,
            "count": len(result)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to generate permutations: {str(e)}"}
