def handler(event):
    items = event.get("items", [])
    index = event.get("index")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if index is None:
        return {"ok": False, "error": "index is required"}

    if not isinstance(index, int):
        return {"ok": False, "error": "index must be an integer"}

    if not items:
        return {"ok": False, "error": "items cannot be empty"}

    if index < 0:
        index = len(items) + index

    if index < 0 or index >= len(items):
        return {"ok": False, "error": f"index {index} is out of bounds for array of length {len(items)}"}

    try:
        # Create a copy of the array
        result = items.copy()

        # Remove the element at the specified index
        deleted_value = result.pop(index)

        return {
            "ok": True,
            "result": result,
            "deleted_value": deleted_value,
            "delete_index": index,
            "original_length": len(items),
            "new_length": len(result)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to delete element: {str(e)}"}
