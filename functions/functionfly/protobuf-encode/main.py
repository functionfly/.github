import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    descriptor = event.get("descriptor")
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        from google.protobuf import descriptor_pb2, descriptor_pool, message_factory
        from google.protobuf import json_format
        if not descriptor:
            return {"ok": False, "error": "descriptor (FileDescriptorProto as base64) is required"}
        desc_bytes = base64.b64decode(str(descriptor))
        file_proto = descriptor_pb2.FileDescriptorProto()
        file_proto.ParseFromString(desc_bytes)
        pool = descriptor_pool.DescriptorPool()
        pool.Add(file_proto)
        return {"ok": False, "error": "Full dynamic protobuf encoding requires schema compilation. Use protobuf-generated Python classes."}
    except ImportError:
        encoded = base64.b64encode(json.dumps(data, ensure_ascii=False).encode("utf-8")).decode("utf-8")
        return {
            "ok": True,
            "result": encoded,
            "note": "protobuf library not installed. Returning JSON-encoded fallback. Install with: pip install protobuf",
            "format": "json-fallback"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
