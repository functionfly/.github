def handler(event):
    """
    Count words in text. Optionally include character and line counts.

    Input:
        - text: String to analyze (required)
        - include_chars: Include character count (default: false)
        - include_lines: Include line count (default: false)

    Returns:
        - ok: True on success
        - words: Word count
        - characters: Present if include_chars
        - lines: Present if include_lines
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        include_chars = event.get("include_chars", False)
        include_lines = event.get("include_lines", False)
    else:
        text = str(event) if event is not None else ""
        include_chars = False
        include_lines = False

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    s = str(text)
    words = len(s.split())
    out = {"ok": True, "words": words}
    if include_chars:
        out["characters"] = len(s)
    if include_lines:
        out["lines"] = len(s.splitlines()) if s.strip() else 0
    return out

