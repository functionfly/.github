import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        import cbor2
        encoded = cbor2.dumps(data)
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded), "encoding": "base64"}
    except ImportError:
        return {"ok": False, "error": "cbor2 library is not installed. Install with: pip install cbor2"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
