import statistics


def handler(event):
    items = event.get("items", [])
    percentiles = event.get("percentiles", [50])  # Default to median

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not items:
        return {"ok": False, "error": "items cannot be empty"}

    if not isinstance(percentiles, list):
        percentiles = [percentiles]

    # Validate percentiles
    for p in percentiles:
        if not isinstance(p, (int, float)) or p < 0 or p > 100:
            return {"ok": False, "error": f"percentiles must be numbers between 0 and 100, got {p}"}

    try:
        # Sort the items
        sorted_items = sorted(items)

        results = []
        for p in percentiles:
            if len(sorted_items) == 1:
                percentile_value = sorted_items[0]
            else:
                # Use linear interpolation method
                k = (len(sorted_items) - 1) * (p / 100)
                f = int(k)
                c = k - f

                if f + 1 < len(sorted_items):
                    percentile_value = sorted_items[f] + c * (sorted_items[f + 1] - sorted_items[f])
                else:
                    percentile_value = sorted_items[f]

            results.append({
                "percentile": p,
                "value": percentile_value
            })

        return {
            "ok": True,
            "result": results,
            "count": len(results)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to calculate percentiles: {str(e)}"}
