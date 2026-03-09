def handler(event):
    arrays = event.get("arrays") if isinstance(event, dict) else None
    if arrays is None:
        return {"ok": False, "error": "arrays is required"}
    if not isinstance(arrays, (list, tuple)):
        return {"ok": False, "error": "arrays must be an array"}
    for i, a in enumerate(arrays):
        if not isinstance(a, (list, tuple)):
            return {"ok": False, "error": f"arrays[{i}] must be an array"}
    if len(arrays) == 0:
        return {"ok": True, "result": []}
    result = [list(t) for t in zip(*arrays)]
    return {"ok": True, "result": result}
