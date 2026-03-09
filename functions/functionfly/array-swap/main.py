def handler(event):
    items = event.get("items", [])
    index1 = event.get("index1")
    index2 = event.get("index2")

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if index1 is None or index2 is None:
        return {"ok": False, "error": "index1 and index2 are required"}

    if not isinstance(index1, int) or not isinstance(index2, int):
        return {"ok": False, "error": "index1 and index2 must be integers"}

    if index1 < 0 or index1 >= len(items):
        return {"ok": False, "error": f"index1 {index1} is out of bounds for array of length {len(items)}"}

    if index2 < 0 or index2 >= len(items):
        return {"ok": False, "error": f"index2 {index2} is out of bounds for array of length {len(items)}"}

    try:
        # Create a copy of the array
        result = items.copy()

        # Swap elements
        result[index1], result[index2] = result[index2], result[index1]

        return {
            "ok": True,
            "result": result,
            "swapped_elements": [items[index1], items[index2]],
            "index1": index1,
            "index2": index2
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to swap elements: {str(e)}"}
