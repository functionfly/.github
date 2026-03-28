import re

def parse_textile(textile: str) -> str:
    """Parse Textile to HTML"""
    html = textile
    # Headers
    html = re.sub(r'^h1\.\s+(.+)$', r'<h1>\1</h1>', html, flags=re.MULTILINE)
    html = re.sub(r'^h2\.\s+(.+)$', r'<h2>\1</h2>', html, flags=re.MULTILINE)
    html = re.sub(r'^h3\.\s+(.+)$', r'<h3>\1</h3>', html, flags=re.MULTILINE)
    html = re.sub(r'^h4\.\s+(.+)$', r'<h4>\1</h4>', html, flags=re.MULTILINE)
    html = re.sub(r'^h5\.\s+(.+)$', r'<h5>\1</h5>', html, flags=re.MULTILINE)
    html = re.sub(r'^h6\.\s+(.+)$', r'<h6>\1</h6>', html, flags=re.MULTILINE)
    # Bold
    html = re.sub(r'\*(.+?)\*', r'<strong>\1</strong>', html)
    # Italic
    html = re.sub(r'_(.+?)_', r'<em>\1</em>', html)
    # Code
    html = re.sub(r'@(.+?)@', r'<code>\1</code>', html)
    # Links
    html = re.sub(r'"(.+?)":(.+?)(?=\s|$)', r'<a href="\2">\1</a>', html)
    # Images
    html = re.sub(r'!(.+?)!(?::(.+?))?(?=\s|$)', r'<img src="\1" alt="\2">', html)
    # Line breaks
    html = re.sub(r'\n\n', r'</p><p>', html)
    html = re.sub(r'\n', r'<br>', html)
    # Wrap in paragraphs
    if not html.startswith('<h') and not html.startswith('<p>'):
        html = f'<p>{html}</p>'
    return html

def handler(event):
    try:
        textile = event.get("textile", "") if isinstance(event, dict) else ""
        if not textile:
            return {"ok": False, "error": "textile is required"}
        html = parse_textile(textile)
        return {"ok": True, "html": html}
    except Exception as e:
        return {"ok": False, "error": str(e)}
