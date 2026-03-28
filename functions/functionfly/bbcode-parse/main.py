import re

def parse_bbcode(bbcode: str) -> str:
    """Parse BBCode to HTML"""
    html = bbcode
    # Bold
    html = re.sub(r'\[b\](.+?)\[/b\]', r'<b>\1</b>', html)
    # Italic
    html = re.sub(r'\[i\](.+?)\[/i\]', r'<i>\1</i>', html)
    # Underline
    html = re.sub(r'\[u\](.+?)\[/u\]', r'<u>\1</u>', html)
    # Strikethrough
    html = re.sub(r'\[s\](.+?)\[/s\]', r'<s>\1</s>', html)
    # Code
    html = re.sub(r'\[code\](.+?)\[/code\]', r'<code>\1</code>', html)
    # Quote
    html = re.sub(r'\[quote\](.+?)\[/quote\]', r'<blockquote>\1</blockquote>', html)
    # URL
    html = re.sub(r'\[url=(.+?)\](.+?)\[/url\]', r'<a href="\1">\2</a>', html)
    # Image
    html = re.sub(r'\[img\](.+?)\[/img\]', r'<img src="\1">', html)
    # Color
    html = re.sub(r'\[color=(.+?)\](.+?)\[/color\]', r'<span style="color:\1">\2</span>', html)
    # Size
    html = re.sub(r'\[size=(.+?)\](.+?)\[/size\]', r'<span style="font-size:\1">\2</span>', html)
    # List
    html = re.sub(r'\[list\](.+?)\[/list\]', r'<ul>\1</ul>', html, flags=re.DOTALL)
    html = re.sub(r'\[\*\](.+?)(?=\[\*\]|\[/list\])', r'<li>\1</li>', html, flags=re.DOTALL)
    return html

def handler(event):
    try:
        bbcode = event.get("bbcode", "") if isinstance(event, dict) else ""
        if not bbcode:
            return {"ok": False, "error": "bbcode is required"}
        html = parse_bbcode(bbcode)
        return {"ok": True, "html": html}
    except Exception as e:
        return {"ok": False, "error": str(e)}
