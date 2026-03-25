import json, base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    mode = event.get("mode", "encode")
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        if mode == "encode":
            encoded = base64.b64encode(json.dumps(data, ensure_ascii=False).encode("utf-8")).decode("utf-8")
            return {
                "ok": True,
                "result": encoded,
                "note": ".NET Binary Serialization format cannot be natively produced without .NET runtime. Returning JSON-encoded fallback.",
                "format": "json-fallback"
            }
        elif mode == "decode":
            raw = base64.b64decode(str(data)).decode("utf-8")
            result = json.loads(raw)
            return {"ok": True, "result": result, "note": "Decoded as JSON fallback"}
        else:
            return {"ok": False, "error": "mode must be 'encode' or 'decode'"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
