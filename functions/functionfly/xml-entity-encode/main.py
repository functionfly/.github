from xml.sax.saxutils import escape


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    entities = event.get("entities", {})

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        result = escape(str(data), entities)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
