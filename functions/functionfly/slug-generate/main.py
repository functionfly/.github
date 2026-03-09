import re


def handler(event):
    """
    Generate URL-friendly slugs from text (same behavior as slugify).
    Lowercase, replace non-alphanumeric with separator, collapse separators.

    Input:
        - text: Text to convert to slug (required)
        - separator: Character between words (default: "-")
        - max_length: Max slug length, 0 = no limit (default: 0)

    Returns:
        - ok: True on success
        - slug: URL-friendly slug
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("input", ""))
        separator = str(event.get("separator", "-")).strip() or "-"
        max_length = event.get("max_length", 0)
        try:
            max_length = int(max_length)
        except (TypeError, ValueError):
            max_length = 0
    else:
        text = str(event)
        separator = "-"
        max_length = 0

    if text is None or (isinstance(text, str) and not text.strip()):
        return {"ok": False, "error": "Input 'text' is required and cannot be empty"}

    s = str(text).lower().strip()
    s = re.sub(r"[^a-z0-9\s-]", "", s)
    s = re.sub(r"[-\s]+", separator, s).strip(separator)
    if max_length > 0:
        s = s[:max_length].rstrip(separator)

    return {"ok": True, "slug": s}
