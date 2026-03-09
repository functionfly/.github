def handler(event):
    items = event.get("items", [])
    count = event.get("count", 1)

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(count, int) or count < 0:
        return {"ok": False, "error": "count must be a non-negative integer"}

    try:
        result = items * count

        return {
            "ok": True,
            "result": result,
            "original_length": len(items),
            "count": count,
            "total_length": len(result)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to repeat array: {str(e)}"}