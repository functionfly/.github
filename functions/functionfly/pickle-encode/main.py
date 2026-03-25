import pickle, base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    protocol = int(event.get("protocol", 5))
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        encoded = pickle.dumps(data, protocol=protocol)
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded), "protocol": protocol}
    except Exception as e:
        return {"ok": False, "error": str(e)}
