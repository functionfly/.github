import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    byte_order = event.get("byte_order", "little")

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        s = str(data)
        enc = "utf-32-le" if byte_order == "little" else "utf-32-be"
        raw = s.encode(enc)
        encoded = base64.b64encode(raw).decode("utf-8")
        return {"ok": True, "result": encoded, "byte_length": len(raw), "byte_order": byte_order}
    except Exception as e:
        return {"ok": False, "error": str(e)}
