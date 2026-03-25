def _to_ron(obj, indent=0):
    pad = "    " * indent
    if isinstance(obj, dict):
        if not obj:
            return "()"
        parts = []
        for k, v in obj.items():
            parts.append(f"{pad}    {k}: {_to_ron(v, indent+1)}")
        return "(\n" + ",\n".join(parts) + f",\n{pad})"
    elif isinstance(obj, list):
        if not obj:
            return "[]"
        parts = [f"{pad}    {_to_ron(v, indent+1)}" for v in obj]
        return "[\n" + ",\n".join(parts) + f",\n{pad}]"
    elif isinstance(obj, bool):
        return "true" if obj else "false"
    elif obj is None:
        return "None"
    elif isinstance(obj, str):
        escaped = obj.replace("\\", "\\\\").replace('"', '\\"')
        return f'"{escaped}"'
    return str(obj)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        result = _to_ron(data)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
