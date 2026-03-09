from collections import Counter


def handler(event):
    items = event.get("items", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    try:
        frequency = Counter(items)

        # Convert to sorted list of [item, count] pairs for consistent output
        result = [[item, count] for item, count in sorted(frequency.items())]

        return {
            "ok": True,
            "result": result,
            "unique_count": len(frequency),
            "total_count": len(items)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to count frequencies: {str(e)}"}
