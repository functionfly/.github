import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    errors = event.get("errors", "strict")

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        raw = str(data).encode("iso-8859-1", errors=errors)
        encoded = base64.b64encode(raw).decode("utf-8")
        return {"ok": True, "result": encoded, "byte_length": len(raw)}
    except (UnicodeEncodeError, LookupError) as e:
        return {"ok": False, "error": str(e)}
