import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded BSON)"}
    try:
        import bson
        result = bson.loads(base64.b64decode(str(data)))
        return {"ok": True, "result": result}
    except ImportError:
        return {"ok": False, "error": "bson library is not installed. Install with: pip install bson"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
