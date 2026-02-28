from urllib.parse import unquote


def handler(event):
    try:
        if isinstance(event, dict):
            text = event.get("text", "")
        else:
            text = str(event) if event is not None else ""

        if text is None:
            return {"ok": False, "error": "Missing required field: text"}

        if not isinstance(text, str):
            text = str(text)

        result = unquote(text)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
