import bisect


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
    i = bisect.bisect_left(arr, value)
    if i < len(arr) and arr[i] == value:
        return {"ok": True, "index": i}
    return {"ok": True, "index": -1}
