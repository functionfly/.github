def handler(event):
    size = event.get("size", 0)
    value = event.get("value", 0)
    start = event.get("start", 0)
    end = event.get("end")

    if not isinstance(size, int) or size < 0:
        return {"ok": False, "error": "size must be a non-negative integer"}

    if end is not None and (not isinstance(end, int) or end < 0):
        return {"ok": False, "error": "end must be a non-negative integer"}

    if not isinstance(start, int) or start < 0:
        return {"ok": False, "error": "start must be a non-negative integer"}

    try:
        # Create array of specified size filled with value
        result = [value] * size

        # If start and end are specified, fill only that range
        if end is not None:
            if start >= size:
                return {"ok": False, "error": "start index must be less than array size"}

            actual_end = min(end, size)
            for i in range(max(start, 0), actual_end):
                result[i] = value

        return {
            "ok": True,
            "result": result,
            "size": len(result),
            "value": value,
            "start": start,
            "end": end
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to fill array: {str(e)}"}