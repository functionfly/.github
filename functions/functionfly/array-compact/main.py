import math


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    result = []
    for x in items:
        if x is None or x is False or x == 0 or x == "":
            continue
        if isinstance(x, float) and math.isnan(x):
            continue
        result.append(x)
    return {"ok": True, "result": result}
