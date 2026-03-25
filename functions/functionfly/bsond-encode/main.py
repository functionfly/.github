import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object (BSON only supports top-level documents)"}
    try:
        import bson
        encoded = bson.dumps(data)
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded)}
    except ImportError:
        try:
            from pymongo.bson import dumps as mongo_dumps
            encoded = mongo_dumps(data)
            return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded)}
        except ImportError:
            return {"ok": False, "error": "bson library is not installed. Install with: pip install bson"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
