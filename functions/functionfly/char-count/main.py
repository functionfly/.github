def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        exclude_whitespace = event.get("exclude_whitespace", False)
    else:
        text = str(event) if event is not None else ""
        exclude_whitespace = False

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text)
    if exclude_whitespace:
        s = "".join(c for c in s if not c.isspace())
    return {"ok": True, "characters": len(s)}

