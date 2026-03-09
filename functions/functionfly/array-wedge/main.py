def handler(event):
    base_array = event.get("base_array", [])
    wedge_array = event.get("wedge_array", [])
    interval = event.get("interval", 1)

    if not isinstance(base_array, list):
        return {"ok": False, "error": "base_array must be an array"}

    if not isinstance(wedge_array, list):
        return {"ok": False, "error": "wedge_array must be an array"}

    if not isinstance(interval, int) or interval <= 0:
        return {"ok": False, "error": "interval must be a positive integer"}

    try:
        result = []
        base_index = 0
        wedge_index = 0

        while base_index < len(base_array) or wedge_index < len(wedge_array):
            # Add elements from base array up to the interval
            for i in range(interval):
                if base_index < len(base_array):
                    result.append(base_array[base_index])
                    base_index += 1
                else:
                    break

            # Add one element from wedge array if available
            if wedge_index < len(wedge_array):
                result.append(wedge_array[wedge_index])
                wedge_index += 1

        return {
            "ok": True,
            "result": result,
            "count": len(result),
            "base_length": len(base_array),
            "wedge_length": len(wedge_array),
            "interval": interval
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to wedge arrays: {str(e)}"}