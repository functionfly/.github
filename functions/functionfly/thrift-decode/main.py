import base64, json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded Thrift or JSON fallback)"}
    try:
        raw = base64.b64decode(str(data)).decode("utf-8")
        result = json.loads(raw)
        return {"ok": True, "result": result, "note": "Decoded as JSON fallback. Full Thrift decoding requires thrift library + generated classes."}
    except Exception as e:
        return {"ok": False, "error": str(e)}
