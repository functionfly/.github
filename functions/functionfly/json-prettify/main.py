import json


def handler(event):
    """
    Pretty-print JSON with indentation.

    Input:
        - data: JSON string or object to prettify (required)
        - indent: Number of spaces for indentation (default: 2)

    Returns:
        - ok: True on success
        - json: Pretty-printed JSON string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        data = event.get("data", event.get("json", event.get("text", event)))
        indent = event.get("indent", 2)
    else:
        data = event
        indent = 2

    if data is None:
        return {"ok": False, "error": "Input 'data' is required"}

    try:
        if isinstance(data, str):
            obj = json.loads(data)
        else:
            obj = data
        try:
            indent = int(indent)
            if indent < 0:
                indent = 0
        except (TypeError, ValueError):
            indent = 2
        out = json.dumps(obj, indent=indent, ensure_ascii=False)
        return {"ok": True, "json": out}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"Invalid JSON: {e!s}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
