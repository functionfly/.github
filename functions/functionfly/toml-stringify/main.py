def _to_toml(obj, indent=0):
    """Simple TOML serializer (no external deps)."""
    lines = []
    prefix = "  " * indent

    def _val(v, key=None):
        if isinstance(v, bool):
            return "true" if v else "false"
        if isinstance(v, int):
            return str(v)
        if isinstance(v, float):
            return str(v)
        if isinstance(v, str):
            escaped = v.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n").replace("\r", "\\r")
            return f'"{escaped}"'
        if isinstance(v, list):
            if all(not isinstance(i, (dict, list)) for i in v):
                items = ", ".join(_val(i) for i in v)
                return f"[{items}]"
            return None
        return None

    simple = {}
    tables = {}
    for k, v in obj.items():
        if isinstance(v, dict):
            tables[k] = v
        else:
            simple[k] = v

    for k, v in simple.items():
        rv = _val(v)
        if rv is not None:
            lines.append(f"{k} = {rv}")

    for k, v in tables.items():
        lines.append(f"\n[{k}]")
        for sk, sv in v.items():
            rv = _val(sv)
            if rv is not None:
                lines.append(f"{sk} = {rv}")

    return "\n".join(lines)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required (object to serialize)"}
    if not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object/dict"}
    try:
        result = _to_toml(data)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
