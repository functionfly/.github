import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        import msgpack
        encoded = msgpack.packb(data, use_bin_type=True)
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded), "encoding": "base64"}
    except ImportError:
        return {"ok": False, "error": "msgpack library is not installed. Install with: pip install msgpack"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
