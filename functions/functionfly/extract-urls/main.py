import re


def handler(event):
    """
    Extract all URLs from text.

    Input:
        - text: String to search (required)

    Returns:
        - ok: True on success
        - urls: List of unique URL strings (order of first occurrence)
        - count: Number of URLs found
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    # Match http(s) and common URL-like patterns
    pattern = r"https?://[^\s<>\"']+|www\.[^\s<>\"']+|[a-zA-Z0-9][-a-zA-Z0-9.]*\.(?:com|org|net|io|co|dev)[^\s<>\"']*"
    found = re.findall(pattern, str(text))
    seen = set()
    urls = []
    for u in found:
        u = u.strip()
        if u and u not in seen:
            if not u.startswith(("http://", "https://")):
                u = "https://" + u
            seen.add(u)
            urls.append(u)
    return {"ok": True, "urls": urls, "count": len(urls)}

