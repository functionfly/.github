def handler(event):
    items = event.get("items", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    try:
        length = len(items)
        mid = length // 2

        first_half = items[:mid]
        second_half = items[mid:]

        return {
            "ok": True,
            "result": [first_half, second_half],
            "first_half": first_half,
            "second_half": second_half,
            "total_length": length,
            "midpoint": mid
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to split array: {str(e)}"}