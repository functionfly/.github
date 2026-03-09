def handler(event):
    """
    Encode a string to hexadecimal.

    Input:
        - text: String to encode (required)

    Returns:
        - ok: True on success
        - encoded: Hex string (e.g. 68656c6c6f)
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = event

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    try:
        data = text.encode("utf-8") if isinstance(text, str) else bytes(text)
        encoded = data.hex()
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
