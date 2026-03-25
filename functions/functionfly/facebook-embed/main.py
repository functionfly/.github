def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    embed_type = event.get("type", "post")
    width = int(event.get("width", 500))
    show_text = event.get("show_text", True)
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        import urllib.parse
        encoded = urllib.parse.quote(str(url), safe='')
        if embed_type == "video":
            embed_url = f"https://www.facebook.com/plugins/video.php?href={encoded}&width={width}&show_text={str(show_text).lower()}"
            height = int(width * 9 / 16) + (50 if show_text else 0)
        else:
            embed_url = f"https://www.facebook.com/plugins/post.php?href={encoded}&width={width}&show_text={str(show_text).lower()}"
            height = 500
        html = f'<iframe src="{embed_url}" width="{width}" height="{height}" style="border:none;overflow:hidden" scrolling="no" frameborder="0" allowfullscreen allow="autoplay; clipboard-write; encrypted-media; picture-in-picture; web-share"></iframe>'
        return {"ok": True, "result": html, "embed_html": html, "embed_url": embed_url}
    except Exception as e:
        return {"ok": False, "error": str(e)}
