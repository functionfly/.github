import plistlib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    fmt = event.get("format", "xml")
    if data is None or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object"}
    try:
        plist_fmt = plistlib.FMT_XML if fmt == "xml" else plistlib.FMT_BINARY
        raw = plistlib.dumps(data, fmt=plist_fmt)
        if fmt == "xml":
            return {"ok": True, "result": raw.decode("utf-8")}
        else:
            import base64
            return {"ok": True, "result": base64.b64encode(raw).decode("utf-8"), "encoding": "base64"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
