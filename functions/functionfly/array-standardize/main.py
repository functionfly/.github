import statistics


def handler(event):
    items = event.get("items", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not items:
        return {"ok": True, "result": [], "count": 0}

    try:
        # Convert to float
        numeric_items = []
        for item in items:
            try:
                numeric_items.append(float(item))
            except (TypeError, ValueError):
                return {"ok": False, "error": f"all items must be numeric, got {type(item)}"}

        # Calculate mean and standard deviation
        mean = statistics.mean(numeric_items)
        stdev = statistics.stdev(numeric_items)

        if stdev == 0:
            # All values are the same, return zeros
            result = [0.0] * len(numeric_items)
        else:
            # Standardize: (x - mean) / std
            result = [(item - mean) / stdev for item in numeric_items]

        return {
            "ok": True,
            "result": result,
            "count": len(result),
            "mean": mean,
            "std": stdev
        }

    except (TypeError, ValueError, statistics.StatisticsError) as e:
        return {"ok": False, "error": f"failed to standardize array: {str(e)}"}
