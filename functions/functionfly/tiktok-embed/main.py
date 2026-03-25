import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    video_id = event.get("video_id")
    width = int(event.get("width", 325))
    height = int(event.get("height", 575))
    if not url and not video_id:
        return {"ok": False, "error": "url or video_id is required"}
    try:
        if not video_id and url:
            m = re.search(r'tiktok\.com/@[^/]+/video/(\d+)', str(url))
            video_id = m.group(1) if m else None
        if video_id:
            embed_url = f"https://www.tiktok.com/embed/v2/{video_id}"
            html = f'<iframe src="{embed_url}" width="{width}" height="{height}" allowfullscreen frameborder="0" scrolling="no" allow="encrypted-media"></iframe>'
            return {"ok": True, "result": html, "embed_html": html, "embed_url": embed_url, "video_id": video_id}
        return {"ok": True, "result": None, "embed_html": None, "note": "Could not extract TikTok video ID from URL"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
