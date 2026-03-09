import re


def handler(event):
    """
    Convert text to camelCase.

    Input:
        - text: String to convert (required)

    Returns:
        - ok: True on success
        - result: camelCase string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text).strip()
    words = re.split(r"[\s_\-]+", s)
    if not words:
        return {"ok": True, "result": ""}
    result = words[0].lower()
    for w in words[1:]:
        if w:
            result += w[0].upper() + w[1:].lower()
    return {"ok": True, "result": result}
