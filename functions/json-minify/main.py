import json


def handler(event):
    try:
        if isinstance(event, dict):
            data = event.get("json", "")
        else:
            data = str(event) if event is not None else ""

        if data is None or (isinstance(data, str) and not data.strip()):
            return {"ok": False, "error": "Missing required field: json"}

        if isinstance(data, str):
            obj = json.loads(data)
        else:
            obj = data

        result = json.dumps(obj, separators=(",", ":"), ensure_ascii=False)
        return {"ok": True, "result": result}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"Invalid JSON: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
