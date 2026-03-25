def _to_hocon(obj, indent=0):
    pad = "  " * indent
    if isinstance(obj, dict):
        lines = ["{"]
        for k, v in obj.items():
            inner = _to_hocon(v, indent+1)
            if isinstance(v, dict):
                lines.append(f"{pad}  {k} {inner}")
            elif isinstance(v, str):
                lines.append(f'{pad}  {k} = "{v}"')
            elif isinstance(v, bool):
                lines.append(f"{pad}  {k} = {'true' if v else 'false'}")
            elif v is None:
                lines.append(f"{pad}  {k} = null")
            else:
                lines.append(f"{pad}  {k} = {v}")
        lines.append(pad + "}")
        return "\n".join(lines)
    elif isinstance(obj, list):
        return "[" + ", ".join(_to_hocon(i) for i in obj) + "]"
    elif isinstance(obj, str):
        return f'"{obj}"'
    elif isinstance(obj, bool):
        return "true" if obj else "false"
    elif obj is None:
        return "null"
    return str(obj)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object"}
    try:
        inner_lines = []
        for k, v in data.items():
            inner = _to_hocon(v, 0)
            if isinstance(v, dict):
                inner_lines.append(f"{k} {inner}")
            elif isinstance(v, str):
                inner_lines.append(f'{k} = "{v}"')
            elif isinstance(v, bool):
                inner_lines.append(f"{k} = {'true' if v else 'false'}")
            elif v is None:
                inner_lines.append(f"{k} = null")
            else:
                inner_lines.append(f"{k} = {v}")
        return {"ok": True, "result": "\n".join(inner_lines)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
