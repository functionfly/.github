import re


def handler(event):
    """
    Extract all email addresses from text.

    Input:
        - text: String to search (required)

    Returns:
        - ok: True on success
        - emails: List of unique email strings (order of first occurrence)
        - count: Number of emails found
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    # Common email pattern (not full RFC; good for extraction)
    pattern = r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"
    found = re.findall(pattern, str(text))
    seen = set()
    emails = []
    for e in found:
        e = e.strip()
        if e and e not in seen:
            seen.add(e)
            emails.append(e)
    return {"ok": True, "emails": emails, "count": len(emails)}

