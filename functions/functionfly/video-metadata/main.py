import re


def _extract_yt_id(url):
    for pat in [r'(?:youtube\.com/watch\?[^"]*v=|youtu\.be/)([A-Za-z0-9_-]{11})', r'youtube\.com/embed/([A-Za-z0-9_-]{11})', r'youtube\.com/shorts/([A-Za-z0-9_-]{11})']:
        m = re.search(pat, url)
        if m: return m.group(1)
    return None

def _extract_vimeo_id(url):
    m = re.search(r'vimeo\.com/(?:video/)?(\d+)', url)
    return m.group(1) if m else None


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        u = str(url)
        yt_id = _extract_yt_id(u)
        if yt_id:
            return {
                "ok": True,
                "result": {"platform": "youtube", "video_id": yt_id},
                "platform": "youtube",
                "video_id": yt_id,
                "embed_url": f"https://www.youtube.com/embed/{yt_id}",
                "watch_url": f"https://www.youtube.com/watch?v={yt_id}",
                "thumbnail_url": f"https://img.youtube.com/vi/{yt_id}/maxresdefault.jpg",
                "oembed_url": f"https://www.youtube.com/oembed?url=https://youtu.be/{yt_id}&format=json"
            }
        vimeo_id = _extract_vimeo_id(u)
        if vimeo_id:
            return {
                "ok": True,
                "result": {"platform": "vimeo", "video_id": vimeo_id},
                "platform": "vimeo",
                "video_id": vimeo_id,
                "embed_url": f"https://player.vimeo.com/video/{vimeo_id}",
                "watch_url": f"https://vimeo.com/{vimeo_id}",
                "oembed_url": f"https://vimeo.com/api/oembed.json?url=https://vimeo.com/{vimeo_id}"
            }
        return {"ok": True, "result": {"platform": "unknown"}, "platform": "unknown", "url": u}
    except Exception as e:
        return {"ok": False, "error": str(e)}
