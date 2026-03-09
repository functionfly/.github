import statistics


def handler(event):
    items = event.get("items", [])
    method = event.get("method", "iqr")  # "iqr", "zscore", "modified_zscore"
    threshold = event.get("threshold")  # method-specific threshold

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if len(items) < 3:
        return {"ok": False, "error": "items must have at least 3 elements for outlier detection"}

    valid_methods = ["iqr", "zscore", "modified_zscore"]
    if method not in valid_methods:
        return {"ok": False, "error": f"method must be one of: {', '.join(valid_methods)}"}

    try:
        # Convert to float and sort
        numeric_items = []
        for item in items:
            try:
                numeric_items.append(float(item))
            except (TypeError, ValueError):
                return {"ok": False, "error": f"all items must be numeric, got {type(item)}"}

        sorted_items = sorted(numeric_items)
        outliers = []
        outlier_indices = []

        if method == "iqr":
            # IQR method: outliers are outside 1.5 * IQR from Q1/Q3
            iqr_threshold = threshold if threshold is not None else 1.5

            q1 = statistics.quantiles(sorted_items, n=4, method="inclusive")[0]
            q3 = statistics.quantiles(sorted_items, n=4, method="inclusive")[2]
            iqr = q3 - q1

            lower_bound = q1 - iqr_threshold * iqr
            upper_bound = q3 + iqr_threshold * iqr

            for i, item in enumerate(items):
                numeric_item = float(item)
                if numeric_item < lower_bound or numeric_item > upper_bound:
                    outliers.append(item)
                    outlier_indices.append(i)

        elif method == "zscore":
            # Z-score method: outliers have |z-score| > threshold
            z_threshold = threshold if threshold is not None else 3.0

            mean = statistics.mean(numeric_items)
            stdev = statistics.stdev(numeric_items)

            if stdev == 0:
                return {"ok": False, "error": "cannot use zscore method: all values are identical"}

            for i, item in enumerate(items):
                numeric_item = float(item)
                z_score = abs((numeric_item - mean) / stdev)
                if z_score > z_threshold:
                    outliers.append(item)
                    outlier_indices.append(i)

        elif method == "modified_zscore":
            # Modified Z-score method: outliers have |modified z-score| > threshold
            mz_threshold = threshold if threshold is not None else 3.5

            sorted_items = sorted(numeric_items)
            median = statistics.median(sorted_items)

            # Calculate MAD (Median Absolute Deviation)
            absolute_deviations = [abs(x - median) for x in numeric_items]
            mad = statistics.median(absolute_deviations)

            if mad == 0:
                return {"ok": False, "error": "cannot use modified_zscore method: MAD is zero"}

            for i, item in enumerate(items):
                numeric_item = float(item)
                modified_z_score = 0.6745 * abs(numeric_item - median) / mad
                if modified_z_score > mz_threshold:
                    outliers.append(item)
                    outlier_indices.append(i)

        return {
            "ok": True,
            "result": outliers,
            "outlier_indices": outlier_indices,
            "count": len(outliers),
            "method": method,
            "threshold": threshold
        }

    except (TypeError, ValueError, statistics.StatisticsError) as e:
        return {"ok": False, "error": f"failed to detect outliers: {str(e)}"}
