import base64

def handler(event):
    try:
        data = event.get("data", "") if isinstance(event, dict) else ""
        if not data:
            return {"ok": False, "error": "data is required"}
        encoded = base64.b85encode(data.encode('utf-8')).decode('utf-8')
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
