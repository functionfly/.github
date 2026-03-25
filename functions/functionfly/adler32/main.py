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
        value = zlib.adler32(raw) & 0xFFFFFFFF
        return {"ok": True, "result": f"{value:08x}" if format_ == "hex" else value, "hex": f"{value:08x}", "decimal": value}
    except Exception as e:
        return {"ok": False, "error": str(e)}
