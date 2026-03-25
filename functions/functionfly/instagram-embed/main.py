import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    width = int(event.get("width", 400))
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        clean = re.sub(r'\?.*$', '', str(url).rstrip('/'))
        embed_url = f"{clean}/embed"
        html = f'<iframe src="{embed_url}" width="{width}" height="{width + 250}" frameborder="0" scrolling="no" allowtransparency></iframe>'
        return {"ok": True, "result": html, "embed_html": html, "embed_url": embed_url}
    except Exception as e:
        return {"ok": False, "error": str(e)}
