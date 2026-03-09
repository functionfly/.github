from collections import defaultdict


def handler(event):
    items = event.get("items", [])
    order = event.get("order", "asc")  # "asc" or "desc"
    method = event.get("method", "standard")  # "standard", "min", "max", "dense", "ordinal"

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if order not in ["asc", "desc"]:
        return {"ok": False, "error": "order must be 'asc' or 'desc'"}

    if method not in ["standard", "min", "max", "dense", "ordinal"]:
        return {"ok": False, "error": f"method must be one of: standard, min, max, dense, ordinal"}

    if not items:
        return {"ok": True, "result": [], "count": 0}

    try:
        # Sort items with their original indices
        indexed_items = [(item, i) for i, item in enumerate(items)]

        if order == "asc":
            indexed_items.sort(key=lambda x: x[0])
        else:  # desc
            indexed_items.sort(key=lambda x: x[0], reverse=True)

        # Assign ranks based on method
        ranks = [0] * len(items)

        if method == "ordinal":
            # Simple 1, 2, 3, ... ranking
            for rank, (_, original_index) in enumerate(indexed_items, 1):
                ranks[original_index] = rank

        elif method == "dense":
            # Dense ranking: 1, 1, 2, 2, 3, ...
            current_rank = 1
            prev_value = None

            for (value, original_index), rank_pos in zip(indexed_items, range(1, len(indexed_items) + 1)):
                if prev_value is None or value != prev_value:
                    current_rank = rank_pos
                    prev_value = value
                ranks[original_index] = current_rank

        elif method == "standard":
            # Standard competition ranking: 1, 2, 2, 4, ...
            current_rank = 1
            group_start = 0

            for i in range(1, len(indexed_items)):
                if indexed_items[i][0] != indexed_items[i-1][0]:
                    # Assign ranks to the previous group
                    for j in range(group_start, i):
                        ranks[indexed_items[j][1]] = current_rank
                    current_rank = i + 1
                    group_start = i

            # Handle the last group
            for j in range(group_start, len(indexed_items)):
                ranks[indexed_items[j][1]] = current_rank

        elif method == "min":
            # Minimum ranking: 1, 2, 2, 2, ...
            current_rank = 1
            group_start = 0

            for i in range(1, len(indexed_items)):
                if indexed_items[i][0] != indexed_items[i-1][0]:
                    # Assign the current rank to all in the previous group
                    for j in range(group_start, i):
                        ranks[indexed_items[j][1]] = current_rank
                    current_rank = i + 1
                    group_start = i

            # Handle the last group
            for j in range(group_start, len(indexed_items)):
                ranks[indexed_items[j][1]] = current_rank

        elif method == "max":
            # Maximum ranking: 3, 3, 3, 6, ...
            current_rank = 1
            group_start = 0

            for i in range(1, len(indexed_items)):
                if indexed_items[i][0] != indexed_items[i-1][0]:
                    # Assign the rank of the end of group to all in the previous group
                    group_rank = i
                    for j in range(group_start, i):
                        ranks[indexed_items[j][1]] = group_rank
                    current_rank = i + 1
                    group_start = i

            # Handle the last group
            group_rank = len(indexed_items)
            for j in range(group_start, len(indexed_items)):
                ranks[indexed_items[j][1]] = group_rank

        return {
            "ok": True,
            "result": ranks,
            "count": len(ranks)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to rank array: {str(e)}"}
