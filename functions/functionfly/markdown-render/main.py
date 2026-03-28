import re

def render_markdown_to_html(markdown: str) -> str:
    """Render Markdown to HTML"""
    html = markdown
    # Headers
    html = re.sub(r'^### (.+)$', r'<h3>\1</h3>', html, flags=re.MULTILINE)
    html = re.sub(r'^## (.+)$', r'<h2>\1</h2>', html, flags=re.MULTILINE)
    html = re.sub(r'^# (.+)$', r'<h1>\1</h1>', html, flags=re.MULTILINE)
    # Bold
    html = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', html)
    # Italic
    html = re.sub(r'\*(.+?)\*', r'<em>\1</em>', html)
    # Code
    html = re.sub(r'`(.+?)`', r'<code>\1</code>', html)
    # Links
    html = re.sub(r'\[(.+?)\]\((.+?)\)', r'<a href="\2">\1</a>', html)
    # Line breaks
    html = re.sub(r'\n\n', r'</p><p>', html)
    html = re.sub(r'\n', r'<br>', html)
    # Wrap in paragraphs
    if not html.startswith('<h') and not html.startswith('<p>'):
        html = f'<p>{html}</p>'
    return html

def handler(event):
    try:
        markdown = event.get("markdown", "") if isinstance(event, dict) else ""
        if not markdown:
            return {"ok": False, "error": "markdown is required"}
        html = render_markdown_to_html(markdown)
        return {"ok": True, "html": html}
    except Exception as e:
        return {"ok": False, "error": str(e)}
