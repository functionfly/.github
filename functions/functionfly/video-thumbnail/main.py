import re


YOUTUBE_PATTERNS = [
    r'(?:youtube\.com/watch\?[^"]*v=|youtu\.be/)([A-Za-z0-9_-]{11})',
    r'youtube\.com/embed/([A-Za-z0-9_-]{11})',
    r'youtube\.com/shorts/([A-Za-z0-9_-]{11})',
]
VIMEO_PATTERN = r'vimeo\.com/(?:video/)?(\d+)'


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    quality = event.get("quality", "hq")
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        u = str(url)
        for pat in YOUTUBE_PATTERNS:
            m = re.search(pat, u)
            if m:
                vid = m.group(1)
                quality_map = {"default": "default", "mq": "mqdefault", "hq": "hqdefault", "sd": "sddefault", "max": "maxresdefault"}
                q_key = quality_map.get(quality, "hqdefault")
                thumb = f"https://img.youtube.com/vi/{vid}/{q_key}.jpg"
                return {"ok": True, "result": thumb, "thumbnail_url": thumb, "platform": "youtube", "video_id": vid}
        m = re.search(VIMEO_PATTERN, u)
        if m:
            vid = m.group(1)
            return {"ok": True, "result": None, "thumbnail_url": None, "platform": "vimeo", "video_id": vid, "note": "Vimeo thumbnails require API access"}
        return {"ok": True, "result": None, "thumbnail_url": None, "platform": "unknown", "note": "URL not recognized"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
