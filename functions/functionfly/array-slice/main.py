def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    start = event.get("start")
    end = event.get("end")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    arr = list(items)
    if start is None:
        start = 0
    if end is None:
        end = len(arr)
    try:
        start = int(start)
        end = int(end)
    except (TypeError, ValueError):
        return {"ok": False, "error": "start and end must be integers"}
    result = arr[start:end]
    return {"ok": True, "result": result}
