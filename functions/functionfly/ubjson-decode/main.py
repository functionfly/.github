import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded UBJSON)"}
    try:
        import ubjson
        result = ubjson.loadb(base64.b64decode(str(data)))
        return {"ok": True, "result": result}
    except ImportError:
        return {"ok": False, "error": "ubjson library is not installed. Install with: pip install py-ubjson"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
