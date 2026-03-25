import html


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    quote = event.get("quote", True)

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        result = html.escape(str(data), quote=quote)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
