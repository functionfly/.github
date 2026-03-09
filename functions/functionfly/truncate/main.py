def handler(event):
    """
    Truncate string to specified length.

    Input:
        - text: String to truncate (required)
        - max_length: Maximum length (required)
        - suffix: Suffix when truncated (default: "...")

    Returns:
        - ok: True on success
        - result: Truncated string
        - truncated: True if string was truncated
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        max_length = event.get("max_length", 0)
        suffix = event.get("suffix", "...")
    else:
        text = str(event) if event is not None else ""
        max_length = 0
        suffix = "..."

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}
    try:
        max_length = int(max_length)
    except (TypeError, ValueError):
        return {"ok": False, "error": "Input 'max_length' must be a positive integer"}

    if max_length < 0:
        return {"ok": False, "error": "Input 'max_length' must be non-negative"}

    s = str(text)
    suffix = str(suffix) if suffix is not None else "..."
    if len(s) <= max_length:
        return {"ok": True, "result": s, "truncated": False}
    if max_length <= len(suffix):
        result = s[:max_length]
        return {"ok": True, "result": result, "truncated": True}
    result = s[: max_length - len(suffix)] + suffix
    return {"ok": True, "result": result, "truncated": True}
