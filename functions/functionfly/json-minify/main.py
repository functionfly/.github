import json


def handler(event):
    """
    Minify JSON by removing whitespace.

    Input:
        - data: JSON string or object to minify (required)

    Returns:
        - ok: True on success
        - json: Minified JSON string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        data = event.get("data", event.get("json", event.get("input")))
    else:
        data = event

    if data is None:
        return {"ok": False, "error": "Input 'data' is required"}

    try:
        if isinstance(data, str):
            obj = json.loads(data)
        else:
            obj = data
        minified = json.dumps(obj, separators=(",", ":"), ensure_ascii=False)
        return {"ok": True, "json": minified}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"Invalid JSON: {e}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
