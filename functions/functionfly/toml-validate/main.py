import tomllib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (TOML string)"}
    try:
        tomllib.loads(str(data))
        return {"ok": True, "result": True, "valid": True, "message": "Valid TOML"}
    except tomllib.TOMLDecodeError as e:
        return {"ok": True, "result": False, "valid": False, "message": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
