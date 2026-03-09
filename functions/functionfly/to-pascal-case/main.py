import re


def handler(event):
    """
    Convert text to PascalCase.

    Input:
        - text: String to convert (required)

    Returns:
        - ok: True on success
        - result: PascalCase string
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
    result = "".join(w[0].upper() + w[1:].lower() for w in words if w)
    return {"ok": True, "result": result}
