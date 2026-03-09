def handler(event):
    items = event.get("items", [])
    method = event.get("method", "minmax")  # "minmax", "zscore", "robust", "l1", "l2"

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not items:
        return {"ok": True, "result": [], "count": 0}

    valid_methods = ["minmax", "zscore", "robust", "l1", "l2"]
    if method not in valid_methods:
        return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}

    try:
        # Convert to float
        numeric_items = []
        for item in items:
            try:
                numeric_items.append(float(item))
            except (TypeError, ValueError):
                return {"ok": False, "error": f"all items must be numeric, got {type(item)}"}

        result = []

        if method == "minmax":
            # Min-Max normalization: (x - min) / (max - min)
            min_val = min(numeric_items)
            max_val = max(numeric_items)

            if min_val == max_val:
                # All values are the same, return zeros
                result = [0.0] * len(numeric_items)
            else:
                for item in numeric_items:
                    normalized = (item - min_val) / (max_val - min_val)
                    result.append(normalized)

        elif method == "zscore":
            # Z-score normalization: (x - mean) / std
            mean = sum(numeric_items) / len(numeric_items)
            variance = sum((x - mean) ** 2 for x in numeric_items) / len(numeric_items)
            std = variance ** 0.5

            if std == 0:
                # All values are the same, return zeros
                result = [0.0] * len(numeric_items)
            else:
                for item in numeric_items:
                    normalized = (item - mean) / std
                    result.append(normalized)

        elif method == "robust":
            # Robust normalization using median and IQR
            import statistics
            sorted_items = sorted(numeric_items)
            median = statistics.median(sorted_items)
            q1 = statistics.quantiles(sorted_items, n=4, method="inclusive")[0]
            q3 = statistics.quantiles(sorted_items, n=4, method="inclusive")[2]
            iqr = q3 - q1

            if iqr == 0:
                # All values are the same, return zeros
                result = [0.0] * len(numeric_items)
            else:
                for item in numeric_items:
                    normalized = (item - median) / iqr
                    result.append(normalized)

        elif method == "l1":
            # L1 normalization: x / ||x||_1
            l1_norm = sum(abs(x) for x in numeric_items)

            if l1_norm == 0:
                # All values are zero, return zeros
                result = [0.0] * len(numeric_items)
            else:
                for item in numeric_items:
                    normalized = item / l1_norm
                    result.append(normalized)

        elif method == "l2":
            # L2 normalization: x / ||x||_2
            l2_norm = sum(x ** 2 for x in numeric_items) ** 0.5

            if l2_norm == 0:
                # All values are zero, return zeros
                result = [0.0] * len(numeric_items)
            else:
                for item in numeric_items:
                    normalized = item / l2_norm
                    result.append(normalized)

        return {
            "ok": True,
            "result": result,
            "count": len(result),
            "method": method
        }

    except (TypeError, ValueError, ZeroDivisionError) as e:
        return {"ok": False, "error": f"failed to normalize array: {str(e)}"}
