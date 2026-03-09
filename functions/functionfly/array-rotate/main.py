def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    n = event.get("n")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
    except (TypeError, ValueError):
        return {"ok": False, "error": "n must be an integer"}
    arr = list(items)
    if not arr:
        return {"ok": True, "result": []}
    n = n % len(arr)
    if n == 0:
        return {"ok": True, "result": arr}
    if n > 0:
        result = arr[-n:] + arr[:-n]
    else:
        result = arr[-n:] + arr[:-n]
    return {"ok": True, "result": result}
