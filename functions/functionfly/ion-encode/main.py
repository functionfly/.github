import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    binary = event.get("binary", False)
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        import amazon.ion.simpleion as ion
        encoded = ion.dumps(data, binary=binary)
        if binary:
            return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "format": "binary", "bytes": len(encoded)}
        return {"ok": True, "result": encoded.decode("utf-8") if isinstance(encoded, bytes) else str(encoded), "format": "text"}
    except ImportError:
        text = json.dumps(data, ensure_ascii=False)
        return {"ok": True, "result": text, "note": "amazon.ion not installed; returning JSON fallback. Install with: pip install amazon.ion"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
