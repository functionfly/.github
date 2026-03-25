from urllib.parse import unquote, parse_qs


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    as_object = event.get("as_object", False)

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        val = str(data)
        if as_object:
            parsed = parse_qs(val, keep_blank_values=True)
            result = {k: v[0] if len(v) == 1 else v for k, v in parsed.items()}
            return {"ok": True, "result": result}
        else:
            result = unquote(val)
            return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
