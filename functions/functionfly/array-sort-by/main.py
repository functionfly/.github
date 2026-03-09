def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    key = event.get("key")
    order = event.get("order", "asc")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    reverse = order == "desc" if isinstance(order, str) else False
    if key and isinstance(key, str):
        def sort_key(x):
            if isinstance(x, dict) and key in x:
                return x[key]
            return None
        try:
            result = sorted(items, key=sort_key, reverse=reverse)
        except TypeError:
            return {"ok": False, "error": "incomparable values for key"}
    else:
        try:
            result = sorted(items, reverse=reverse)
        except TypeError as e:
            return {"ok": False, "error": str(e)}
    return {"ok": True, "result": list(result)}
