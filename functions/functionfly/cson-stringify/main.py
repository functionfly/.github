def _to_cson(obj, indent=0):
    lines = []
    pad = "  " * indent
    if isinstance(obj, dict):
        for k, v in obj.items():
            if isinstance(v, dict):
                lines.append(f"{pad}{k}:")
                lines.append(_to_cson(v, indent+1))
            elif isinstance(v, str):
                escaped = v.replace("'", "\\'")
                lines.append(f"{pad}{k}: '{escaped}'")
            elif isinstance(v, bool):
                lines.append(f"{pad}{k}: {'true' if v else 'false'}")
            elif v is None:
                lines.append(f"{pad}{k}: null")
            else:
                lines.append(f"{pad}{k}: {v}")
    return "\n".join(lines)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object"}
    try:
        result = _to_cson(data)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
