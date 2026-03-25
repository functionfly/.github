import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        patterns = [
            r'(?:youtube\.com/watch\?(?:[^&]*&)*v=)([A-Za-z0-9_-]{11})',
            r'(?:youtu\.be/)([A-Za-z0-9_-]{11})',
            r'(?:youtube\.com/embed/)([A-Za-z0-9_-]{11})',
            r'(?:youtube\.com/v/)([A-Za-z0-9_-]{11})',
            r'(?:youtube\.com/shorts/)([A-Za-z0-9_-]{11})',
        ]
        for pattern in patterns:
            m = re.search(pattern, str(url))
            if m:
                vid_id = m.group(1)
                return {
                    "ok": True,
                    "result": vid_id,
                    "video_id": vid_id,
                    "embed_url": f"https://www.youtube.com/embed/{vid_id}",
                    "watch_url": f"https://www.youtube.com/watch?v={vid_id}",
                    "thumbnail_url": f"https://img.youtube.com/vi/{vid_id}/hqdefault.jpg"
                }
        return {"ok": True, "result": None, "video_id": None, "found": False}
    except Exception as e:
        return {"ok": False, "error": str(e)}
