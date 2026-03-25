import zlib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    format_ = event.get("format", "decimal")

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")
        value = zlib.crc32(raw) & 0xFFFFFFFF
        if format_ == "hex":
            result = f"{value:08x}"
        elif format_ == "unsigned":
            result = value
        else:
            result = value
        return {"ok": True, "result": result, "hex": f"{value:08x}", "decimal": value}
    except Exception as e:
        return {"ok": False, "error": str(e)}
