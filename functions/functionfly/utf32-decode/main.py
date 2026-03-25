import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    byte_order = event.get("byte_order", "little")

    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        raw = base64.b64decode(str(data))
        enc = "utf-32-le" if byte_order == "little" else "utf-32-be"
        result = raw.decode(enc)
        return {"ok": True, "result": result, "byte_order": byte_order}
    except Exception as e:
        return {"ok": False, "error": str(e)}
