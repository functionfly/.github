import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded MessagePack bytes)"}
    try:
        import msgpack
        raw = base64.b64decode(str(data))
        result = msgpack.unpackb(raw, raw=False)
        return {"ok": True, "result": result}
    except ImportError:
        return {"ok": False, "error": "msgpack library is not installed. Install with: pip install msgpack"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
