import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    errors = event.get("errors", "strict")

    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        raw = base64.b64decode(str(data))
        result = raw.decode("iso-8859-1", errors=errors)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
