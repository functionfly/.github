import re


def handler(event):
    """
    Convert HTML to Markdown (basic conversion using stdlib).

    Input:
        - html: HTML string to convert (required)

    Returns:
        - ok: True on success
        - markdown: Converted Markdown string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        html = event.get("html", event.get("data", ""))
    else:
        html = str(event) if event is not None else ""

    if html is None or (isinstance(html, str) and not html.strip()):
        return {"ok": False, "error": "Input 'html' is required"}

    text = str(html).strip()

    # Remove script and style elements
    text = re.sub(r"<script[^>]*>[\s\S]*?</script>", "", text, flags=re.IGNORECASE)
    text = re.sub(r"<style[^>]*>[\s\S]*?</style>", "", text, flags=re.IGNORECASE)

    # Headers (h1-h6)
    def _h_repl(m):
        level = len(m.group(1))
        return "\n" + "#" * level + " " + _strip_tags(m.group(2)).strip() + "\n"

    text = re.sub(r"<h([1-6])[^>]*>([\s\S]*?)</h\1>", _h_repl, text, flags=re.IGNORECASE)

    # Links: <a href="url">text</a> -> [text](url)
    text = re.sub(
        r'<a\s+href=["\']([^"\']*)["\'][^>]*>([\s\S]*?)</a>',
        lambda m: "[" + _strip_tags(m.group(2)).strip() + "](" + m.group(1) + ")",
        text,
        flags=re.IGNORECASE,
    )

    # Bold and italic
    text = re.sub(r"<strong[^>]*>([\s\S]*?)</strong>", r"**\1**", text, flags=re.IGNORECASE)
    text = re.sub(r"<b[^>]*>([\s\S]*?)</b>", r"**\1**", text, flags=re.IGNORECASE)
    text = re.sub(r"<em[^>]*>([\s\S]*?)</em>", r"*\1*", text, flags=re.IGNORECASE)
    text = re.sub(r"<i[^>]*>([\s\S]*?)</i>", r"*\1*", text, flags=re.IGNORECASE)

    # Block elements to newlines
    text = re.sub(r"</p>", "\n\n", text, flags=re.IGNORECASE)
    text = re.sub(r"<br\s*/?>", "\n", text, flags=re.IGNORECASE)
    text = re.sub(r"</div>", "\n", text, flags=re.IGNORECASE)
    text = re.sub(r"</li>", "\n", text, flags=re.IGNORECASE)
    text = re.sub(r"<li[^>]*>", "- ", text, flags=re.IGNORECASE)

    # Strip remaining tags
    text = _strip_tags(text)

    # Normalize whitespace
    text = re.sub(r"\n{3,}", "\n\n", text)
    text = text.strip()

    return {"ok": True, "markdown": text}


def _strip_tags(html):
    return re.sub(r"<[^>]+>", "", html)
