import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        import ubjson
        encoded = ubjson.dumpb(data)
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded)}
    except ImportError:
        return {"ok": False, "error": "ubjson library is not installed. Install with: pip install py-ubjson"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
