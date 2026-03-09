from collections import Counter


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    key = event.get("key")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if key:
        if not isinstance(key, str):
            return {"ok": False, "error": "key must be a string"}
        vals = []
        for x in items:
            if isinstance(x, dict) and key in x:
                vals.append(x[key])
            else:
                vals.append(None)
        counts = Counter(vals)
    else:
        counts = Counter(items)
    result = dict(counts)
    return {"ok": True, "result": result}
