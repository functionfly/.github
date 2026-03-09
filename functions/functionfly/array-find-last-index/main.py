def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    value = event.get("value")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if "value" not in event:
        return {"ok": False, "error": "value is required"}
    arr = list(items)
    index = -1
    for i in range(len(arr) - 1, -1, -1):
        if arr[i] == value:
            index = i
            break
    return {"ok": True, "index": index}
