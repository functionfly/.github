def handler(event):
    items = event.get("items", [])
    from_index = event.get("from_index")
    to_index = event.get("to_index")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if from_index is None or to_index is None:
        return {"ok": False, "error": "from_index and to_index are required"}

    if not isinstance(from_index, int) or not isinstance(to_index, int):
        return {"ok": False, "error": "from_index and to_index must be integers"}

    if from_index < 0 or from_index >= len(items):
        return {"ok": False, "error": f"from_index {from_index} is out of bounds for array of length {len(items)}"}

    if to_index < 0 or to_index >= len(items):
        return {"ok": False, "error": f"to_index {to_index} is out of bounds for array of length {len(items)}"}

    try:
        # Create a copy of the array
        result = items.copy()

        # Remove element from from_index
        element = result.pop(from_index)

        # Insert element at to_index
        result.insert(to_index, element)

        return {
            "ok": True,
            "result": result,
            "moved_element": element,
            "from_index": from_index,
            "to_index": to_index
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to move element: {str(e)}"}
