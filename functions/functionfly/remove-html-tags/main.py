import re


def handler(event):
    """
    Strip all HTML tags from text.

    Input:
        - text: String that may contain HTML (required)
        - normalize_whitespace: Collapse whitespace to single spaces (default: true)

    Returns:
        - ok: True on success
        - text: Plain text with tags removed
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        normalize_whitespace = event.get("normalize_whitespace", True)
    else:
        text = str(event) if event is not None else ""
        normalize_whitespace = True

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text)
    s = re.sub(r"<[^>]+>", "", s)
    if normalize_whitespace:
        s = re.sub(r"\s+", " ", s).strip()
    return {"ok": True, "text": s}

