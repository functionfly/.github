import statistics


def handler(event):
    items = event.get("items", [])
    method = event.get("method", "linear")  # "linear", "lower", "higher", "midpoint", "nearest"

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not items:
        return {"ok": False, "error": "items cannot be empty"}

    valid_methods = ["linear", "lower", "higher", "midpoint", "nearest"]
    if method not in valid_methods:
        return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}

    try:
        # Sort the items
        sorted_items = sorted(items)

        # Calculate quartiles using the specified method
        if method == "linear":
            q1 = statistics.quantiles(sorted_items, n=4, method="inclusive")[0]
            q2 = statistics.quantiles(sorted_items, n=4, method="inclusive")[1]
            q3 = statistics.quantiles(sorted_items, n=4, method="inclusive")[2]
        elif method == "lower":
            q1 = statistics.quantiles(sorted_items, n=4, method="lower")[0]
            q2 = statistics.quantiles(sorted_items, n=4, method="lower")[1]
            q3 = statistics.quantiles(sorted_items, n=4, method="lower")[2]
        elif method == "higher":
            q1 = statistics.quantiles(sorted_items, n=4, method="higher")[0]
            q2 = statistics.quantiles(sorted_items, n=4, method="higher")[1]
            q3 = statistics.quantiles(sorted_items, n=4, method="higher")[2]
        elif method == "midpoint":
            q1 = statistics.quantiles(sorted_items, n=4, method="midpoint")[0]
            q2 = statistics.quantiles(sorted_items, n=4, method="midpoint")[1]
            q3 = statistics.quantiles(sorted_items, n=4, method="midpoint")[2]
        elif method == "nearest":
            q1 = statistics.quantiles(sorted_items, n=4, method="nearest")[0]
            q2 = statistics.quantiles(sorted_items, n=4, method="nearest")[1]
            q3 = statistics.quantiles(sorted_items, n=4, method="nearest")[2]

        result = {
            "q0": sorted_items[0],  # minimum
            "q1": q1,  # first quartile (25th percentile)
            "q2": q2,  # second quartile (50th percentile/median)
            "q3": q3,  # third quartile (75th percentile)
            "q4": sorted_items[-1],  # maximum
            "iqr": q3 - q1  # interquartile range
        }

        return {
            "ok": True,
            "result": result,
            "method": method
        }

    except (TypeError, ValueError, statistics.StatisticsError) as e:
        return {"ok": False, "error": f"failed to calculate quartiles: {str(e)}"}
