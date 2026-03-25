import quopri


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    header = event.get("header", False)

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        raw = str(data).encode("utf-8")
        encoded_bytes = quopri.encodestring(raw, header=header)
        result = encoded_bytes.decode("ascii")
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
