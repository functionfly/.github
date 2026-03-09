import json


def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    try:
        result = json.dumps(str(text))
        return {"ok": True, "result": result[1:-1]}
    except Exception as e:
        return {"ok": False, "error": str(e)}

