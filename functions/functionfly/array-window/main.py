from collections import deque


def handler(event):
    items = event.get("items", [])
    size = event.get("size", 2)
    step = event.get("step", 1)

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(size, int) or size <= 0:
        return {"ok": False, "error": "size must be a positive integer"}

    if not isinstance(step, int) or step <= 0:
        return {"ok": False, "error": "step must be a positive integer"}

    if len(items) < size:
        return {"ok": False, "error": f"array length ({len(items)}) must be at least window size ({size})"}

    try:
        windows = []
        for i in range(0, len(items) - size + 1, step):
            window = items[i:i + size]
            windows.append(window)

        return {
            "ok": True,
            "result": windows,
            "count": len(windows)
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"failed to create sliding windows: {str(e)}"}
