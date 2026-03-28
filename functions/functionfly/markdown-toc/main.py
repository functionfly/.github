import re

def generate_toc(markdown: str) -> str:
    """Generate table of contents from Markdown"""
    lines = markdown.split('\n')
    toc_lines = []
    for line in lines:
        match = re.match(r'^(#{1,6})\s+(.+)$', line)
        if match:
            level = len(match.group(1))
            title = match.group(2)
            anchor = re.sub(r'[^\w\s-]', '', title.lower())
            anchor = re.sub(r'[\s]+', '-', anchor)
            indent = '  ' * (level - 1)
            toc_lines.append(f'{indent}- [{title}](#{anchor})')
    return '\n'.join(toc_lines)

def handler(event):
    try:
        markdown = event.get("markdown", "") if isinstance(event, dict) else ""
        if not markdown:
            return {"ok": False, "error": "markdown is required"}
        toc = generate_toc(markdown)
        return {"ok": True, "toc": toc}
    except Exception as e:
        return {"ok": False, "error": str(e)}
