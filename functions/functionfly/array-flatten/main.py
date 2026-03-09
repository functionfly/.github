def _flatten(items, depth):
    out = []
    for x in items:
        if isinstance(x, (list, tuple)) and depth != 0:
            if depth == 1:
                out.extend(x)
            else:
                out.extend(_flatten(x, depth - 1 if depth > 0 else -1))
        else:
            out.append(x)
    return out


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    depth = event.get("depth", 1)
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    try:
        depth = int(depth)
    except (TypeError, ValueError):
        depth = 1
    if depth == 0:
        return {"ok": True, "result": list(items)}
    result = _flatten(list(items), depth)
    return {"ok": True, "result": result}
