import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded protobuf bytes)"}
    try:
        from google.protobuf import any_pb2
        raw = base64.b64decode(str(data))
        msg = any_pb2.Any()
        msg.ParseFromString(raw)
        return {"ok": True, "result": {"type_url": msg.type_url, "value_b64": base64.b64encode(msg.value).decode()}}
    except ImportError:
        try:
            decoded = base64.b64decode(str(data)).decode("utf-8")
            result = json.loads(decoded)
            return {"ok": True, "result": result, "note": "protobuf not installed; decoded as JSON fallback", "format": "json-fallback"}
        except Exception:
            return {"ok": False, "error": "protobuf library is not installed. Install with: pip install protobuf"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
