import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    bom = event.get("bom", True)
    byte_order = event.get("byte_order", "little")

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        s = str(data)
        if byte_order == "big":
            encoding = "utf-16-be"
            if bom:
                raw = b'\xfe\xff' + s.encode("utf-16-be")
            else:
                raw = s.encode("utf-16-be")
        else:
            encoding = "utf-16-le"
            if bom:
                raw = b'\xff\xfe' + s.encode("utf-16-le")
            else:
                raw = s.encode("utf-16-le")

        encoded = base64.b64encode(raw).decode("utf-8")
        return {"ok": True, "result": encoded, "byte_length": len(raw), "byte_order": byte_order, "bom": bom}
    except Exception as e:
        return {"ok": False, "error": str(e)}
