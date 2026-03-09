import re


def handler(event):
    """
    Remove leading and trailing whitespace from a string.

    Input:
        - text: String to trim (required)
        - collapse: If true, collapse internal whitespace to single space (default: false)

    Returns:
        - ok: True on success
        - trimmed: Trimmed string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        collapse = event.get("collapse", False)
    else:
        text = str(event) if event is not None else ""
        collapse = False

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text)
    if collapse:
        s = re.sub(r"\s+", " ", s)
    return {"ok": True, "trimmed": s.strip()}
