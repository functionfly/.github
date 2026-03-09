from collections import defaultdict


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    key = event.get("key")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if not key:
        return {"ok": False, "error": "key is required"}
    if not isinstance(key, str):
        return {"ok": False, "error": "key must be a string"}
    groups = defaultdict(list)
    for x in items:
        if isinstance(x, dict) and key in x:
            k = x[key]
        else:
            k = None
        try:
            groups[k].append(x)
        except TypeError:
            k = str(k) if k is not None else "__null__"
            groups[k].append(x)
    result = dict(groups)
    if "__null__" in result:
        result[None] = result.pop("__null__")
    return {"ok": True, "result": result}
