import plistlib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    fmt = event.get("format", "auto")
    if not data:
        return {"ok": False, "error": "data is required (plist string or Base64 binary plist)"}
    try:
        raw = str(data).encode("utf-8")
        if raw.lstrip().startswith(b"bplist"):
            import base64
            raw = base64.b64decode(data)
        result = plistlib.loads(raw)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
