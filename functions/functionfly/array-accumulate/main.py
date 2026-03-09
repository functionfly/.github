from functools import reduce
import operator


def handler(event):
    items = event.get("items", [])
    operation = event.get("operation", "sum")
    initial = event.get("initial")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not items:
        return {"ok": True, "result": [], "count": 0}

    # Validate operation
    valid_operations = ["sum", "product", "min", "max", "concat"]
    if operation not in valid_operations:
        return {"ok": False, "error": f"operation must be one of: {', '.join(valid_operations)}"}

    try:
        accumulated = []
        current = initial

        for i, item in enumerate(items):
            if current is None:
                current = item
            else:
                if operation == "sum":
                    current = current + item
                elif operation == "product":
                    current = current * item
                elif operation == "min":
                    current = min(current, item)
                elif operation == "max":
                    current = max(current, item)
                elif operation == "concat":
                    if isinstance(current, list) and isinstance(item, list):
                        current = current + item
                    else:
                        current = [current, item] if i == 0 else current + [item]

            accumulated.append(current)

        return {
            "ok": True,
            "result": accumulated,
            "count": len(accumulated)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to accumulate array: {str(e)}"}
