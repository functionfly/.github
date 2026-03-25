import quopri


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    header = event.get("header", False)
    encoding = event.get("encoding", "utf-8")

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        raw = str(data).encode("ascii", errors="replace")
        decoded_bytes = quopri.decodestring(raw, header=header)
        result = decoded_bytes.decode(encoding)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
