import json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not isinstance(data, list):
        return {"ok": False, "error": "data must be an array of objects/values"}
    try:
        lines = [json.dumps(item, ensure_ascii=False, separators=(",", ":")) for item in data]
        return {"ok": True, "result": "\n".join(lines), "count": len(lines)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
