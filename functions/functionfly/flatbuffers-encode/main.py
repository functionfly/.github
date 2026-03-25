import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        import flatbuffers
        return {
            "ok": True,
            "result": base64.b64encode(json.dumps(data).encode()).decode(),
            "note": "FlatBuffers encoding requires schema-generated Python classes. Returning JSON-encoded fallback. Install with: pip install flatbuffers",
            "format": "json-fallback"
        }
    except ImportError:
        return {
            "ok": True,
            "result": base64.b64encode(json.dumps(data).encode()).decode(),
            "note": "flatbuffers not installed; returning JSON fallback. Install with: pip install flatbuffers",
            "format": "json-fallback"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
