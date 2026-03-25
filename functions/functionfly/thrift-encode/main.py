import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    protocol = event.get("protocol", "binary")
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        from thrift.protocol import TBinaryProtocol, TJSONProtocol, TCompactProtocol
        from thrift.transport import TTransport
        encoded = base64.b64encode(json.dumps(data, ensure_ascii=False).encode("utf-8")).decode("utf-8")
        return {
            "ok": True,
            "result": encoded,
            "note": "Full Thrift encoding requires generated classes from a .thrift IDL file. Returning JSON fallback.",
            "protocol": protocol,
            "format": "json-fallback"
        }
    except ImportError:
        encoded = base64.b64encode(json.dumps(data, ensure_ascii=False).encode("utf-8")).decode("utf-8")
        return {
            "ok": True,
            "result": encoded,
            "note": "thrift not installed; returning JSON fallback. Install with: pip install thrift",
            "format": "json-fallback"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
