import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    is_base64 = event.get("is_base64", False)

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        if is_base64:
            raw = base64.b64decode(str(data))
        else:
            raw = str(data).encode("utf-8")

        raw.decode("utf-8")
        return {"ok": True, "result": True, "byte_length": len(raw)}
    except (UnicodeDecodeError, ValueError) as e:
        return {"ok": True, "result": False, "error": str(e)}
