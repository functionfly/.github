import base64


def handler(event):
    """
    Encode a string to Base64.

    Input:
        - text: String to encode (required). Can be plain string or UTF-8 bytes.
        - url_safe: If true, use URL-safe alphabet (default: false)

    Returns:
        - ok: True on success
        - encoded: Base64 string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        url_safe = event.get("url_safe", False)
    else:
        text = str(event) if event is not None else ""
        url_safe = False

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    try:
        if isinstance(text, str):
            data = text.encode("utf-8")
        else:
            data = text
        if url_safe:
            encoded = base64.urlsafe_b64encode(data).decode("ascii").rstrip("=")
        else:
            encoded = base64.b64encode(data).decode("ascii")
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
