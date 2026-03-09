import re


def handler(event):
    """
    Convert text to kebab-case.

    Input:
        - text: String to convert (required)

    Returns:
        - ok: True on success
        - result: kebab-case string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text).strip()
    s = re.sub(r"(?<=[a-z])(?=[A-Z])", "-", s)
    s = re.sub(r"(?<=[A-Z])(?=[A-Z][a-z])", "-", s)
    s = re.sub(r"[\s_]+", "-", s)
    s = s.lower().strip("-")
    s = re.sub(r"-+", "-", s)
    return {"ok": True, "result": s}
