import itertools


def handler(event):
    items = event.get("items", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    try:
        # Generate power set using itertools.combinations
        power_set = []
        for r in range(len(items) + 1):
            combinations = list(itertools.combinations(items, r))
            # Convert tuples back to lists and sort for consistent output
            power_set.extend([list(c) for c in combinations])

        # Sort the power set by subset length, then lexicographically
        power_set.sort(key=lambda x: (len(x), x))

        return {
            "ok": True,
            "result": power_set,
            "count": len(power_set)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to generate power set: {str(e)}"}
