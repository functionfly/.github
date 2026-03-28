import base64

def handler(event):
    try:
        encoded = event.get("encoded", "") if isinstance(event, dict) else ""
        if not encoded:
            return {"ok": False, "error": "encoded is required"}
        data = base64.b32decode(encoded).decode('utf-8')
        return {"ok": True, "data": data}
    except Exception as e:
        return {"ok": False, "error": str(e)}
