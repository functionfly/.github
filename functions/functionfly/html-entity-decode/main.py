import html


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        result = html.unescape(str(data))
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
