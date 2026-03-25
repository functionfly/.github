import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        patterns = [
            r'vimeo\.com/(\d+)',
            r'vimeo\.com/video/(\d+)',
            r'player\.vimeo\.com/video/(\d+)',
        ]
        for pattern in patterns:
            m = re.search(pattern, str(url))
            if m:
                vid_id = m.group(1)
                return {
                    "ok": True,
                    "result": vid_id,
                    "video_id": vid_id,
                    "embed_url": f"https://player.vimeo.com/video/{vid_id}",
                    "watch_url": f"https://vimeo.com/{vid_id}"
                }
        return {"ok": True, "result": None, "video_id": None, "found": False}
    except Exception as e:
        return {"ok": False, "error": str(e)}
