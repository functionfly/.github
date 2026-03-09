def handler(event):
    """
    Convert text to Title Case.

    Input:
        - text: String to convert (required)

    Returns:
        - ok: True on success
        - result: Title-cased string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    result = str(text).title()
    return {"ok": True, "result": result}
