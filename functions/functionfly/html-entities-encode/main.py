import html


def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("content", ""))
        quote = event.get("quote", True)
    else:
        text, quote = "", True

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = text if isinstance(text, str) else str(text)
    encoded = html.escape(s, quote=quote)
    return {"ok": True, "encoded": encoded}
