import json, base64, struct


def _parse_java_serial_header(raw):
    """Parse Java serialization stream magic (0xACED 0x0005) header."""
    if len(raw) < 4:
        return None
    magic = struct.unpack(">H", raw[0:2])[0]
    version = struct.unpack(">H", raw[2:4])[0]
    return {"magic": hex(magic), "version": version, "valid_header": magic == 0xACED and version == 5}


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    mode = event.get("mode", "inspect")
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        if mode == "inspect":
            raw = base64.b64decode(str(data))
            header = _parse_java_serial_header(raw)
            return {
                "ok": True,
                "result": header or {"raw_bytes": len(raw)},
                "bytes": len(raw),
                "note": "Full Java deserialization requires JVM. This function inspects the binary header only."
            }
        elif mode == "encode":
            encoded = base64.b64encode(json.dumps(data, ensure_ascii=False).encode("utf-8")).decode("utf-8")
            return {
                "ok": True,
                "result": encoded,
                "note": "Java Serialization requires JVM. Returning JSON-encoded fallback.",
                "format": "json-fallback"
            }
        else:
            return {"ok": False, "error": "mode must be 'encode' or 'inspect'"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
