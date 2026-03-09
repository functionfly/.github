def handler(event):
    """
    Reverse a string.

    Input:
        - text: String to reverse (required)

    Returns:
        - ok: True on success
        - result: Reversed string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    return {"ok": True, "result": str(text)[::-1]}

