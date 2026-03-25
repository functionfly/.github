import pickle, base64


SAFE_TYPES = (dict, list, str, int, float, bool, tuple, type(None), bytes)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    allow_unsafe = event.get("allow_unsafe", False)
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded pickle)"}
    try:
        raw = base64.b64decode(str(data))
        result = pickle.loads(raw)
        if not allow_unsafe and not isinstance(result, SAFE_TYPES):
            return {"ok": False, "error": f"decoded object type {type(result).__name__!r} is not a safe type. Set allow_unsafe=true to override."}
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
