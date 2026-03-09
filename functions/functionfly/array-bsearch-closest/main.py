def handler(event):
    items = event.get("items", [])
    target = event.get("target")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if target is None:
        return {"ok": False, "error": "target is required"}

    if not items:
        return {"ok": False, "error": "items cannot be empty"}

    try:
        # Convert items to comparable values if needed
        numeric_items = []
        for item in items:
            try:
                numeric_items.append(float(item))
            except (TypeError, ValueError):
                return {"ok": False, "error": f"all items must be numeric, got {type(item)}"}

        target_float = float(target)

        # Binary search for closest element
        left, right = 0, len(numeric_items) - 1
        closest_index = 0
        min_diff = abs(numeric_items[0] - target_float)

        while left <= right:
            mid = (left + right) // 2
            current_diff = abs(numeric_items[mid] - target_float)

            if current_diff < min_diff:
                min_diff = current_diff
                closest_index = mid

            if numeric_items[mid] < target_float:
                left = mid + 1
            elif numeric_items[mid] > target_float:
                right = mid - 1
            else:
                # Exact match found
                closest_index = mid
                break

        # Check adjacent elements for potentially closer values
        if closest_index > 0:
            left_diff = abs(numeric_items[closest_index - 1] - target_float)
            if left_diff < min_diff:
                closest_index = closest_index - 1
                min_diff = left_diff

        if closest_index < len(numeric_items) - 1:
            right_diff = abs(numeric_items[closest_index + 1] - target_float)
            if right_diff < min_diff:
                closest_index = closest_index + 1
                min_diff = right_diff

        return {
            "ok": True,
            "result": {
                "index": closest_index,
                "value": items[closest_index],
                "target": target,
                "difference": abs(numeric_items[closest_index] - target_float)
            },
            "exact_match": numeric_items[closest_index] == target_float
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to find closest element: {str(e)}"}
