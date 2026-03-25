import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    binary = event.get("binary", False)
    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        import amazon.ion.simpleion as ion
        if binary:
            raw = base64.b64decode(str(data))
        else:
            raw = str(data)
        result = ion.loads(raw)
        return {"ok": True, "result": result}
    except ImportError:
        try:
            raw_str = base64.b64decode(str(data)).decode("utf-8") if binary else str(data)
            result = json.loads(raw_str)
            return {"ok": True, "result": result, "note": "amazon.ion not installed; decoded as JSON fallback"}
        except Exception:
            return {"ok": False, "error": "amazon.ion not installed. Install with: pip install amazon.ion"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
