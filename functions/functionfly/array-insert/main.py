def handler(event):
    items = event.get("items", [])
    index = event.get("index", len(items))
    value = event.get("value")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(index, int):
        return {"ok": False, "error": "index must be an integer"}

    if index < 0:
        index = max(0, len(items) + index + 1)

    if index > len(items):
        index = len(items)

    try:
        # Create a copy of the array
        result = items.copy()

        # Insert the value at the specified index
        result.insert(index, value)

        return {
            "ok": True,
            "result": result,
            "inserted_value": value,
            "insert_index": index,
            "original_length": len(items),
            "new_length": len(result)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to insert element: {str(e)}"}
