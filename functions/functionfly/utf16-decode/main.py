import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    byte_order = event.get("byte_order", "auto")

    if not data:
        return {"ok": False, "error": "data is required"}

    try:
        raw = base64.b64decode(str(data))
        if byte_order == "big":
            result = raw.decode("utf-16-be")
        elif byte_order == "little":
            result = raw.decode("utf-16-le")
        else:
            # Auto-detect BOM
            if raw[:2] == b'\xff\xfe':
                result = raw[2:].decode("utf-16-le")
                byte_order = "little"
            elif raw[:2] == b'\xfe\xff':
                result = raw[2:].decode("utf-16-be")
                byte_order = "big"
            else:
                result = raw.decode("utf-16")
                byte_order = "little"
        return {"ok": True, "result": result, "byte_order": byte_order}
    except Exception as e:
        return {"ok": False, "error": str(e)}
