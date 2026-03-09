import urllib.parse


def handler(event):
    """
    URL-encode a string (percent-encoding).

    Input:
        - text: String to encode (required)
        - safe: Optional extra characters to leave unencoded (default: none)

    Returns:
        - ok: True on success
        - encoded: URL-encoded string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", event.get("string", "")))
        safe = event.get("safe", "")
    else:
        text = str(event) if event is not None else ""
        safe = ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    try:
        encoded = urllib.parse.quote(str(text), safe=safe or "")
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
