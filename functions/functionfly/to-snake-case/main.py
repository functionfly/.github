import re


def handler(event):
    """
    Convert text to snake_case.

    Input:
        - text: String to convert (required)

    Returns:
        - ok: True on success
        - result: snake_case string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text).strip()
    s = re.sub(r"[\s\-]+", "_", s)
    s = re.sub(r"(?<=[a-z])(?=[A-Z])", "_", s)
    s = re.sub(r"(?<=[A-Z])(?=[A-Z][a-z])", "_", s)
    s = s.lower().strip("_")
    s = re.sub(r"_+", "_", s)
    return {"ok": True, "result": s}
