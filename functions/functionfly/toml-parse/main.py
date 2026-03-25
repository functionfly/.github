import tomllib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (TOML string)"}
    try:
        result = tomllib.loads(str(data))
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
