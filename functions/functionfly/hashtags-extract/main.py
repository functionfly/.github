import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    platform = event.get("platform", "twitter")
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        if platform in ("twitter", "instagram", "tiktok"):
            pattern = r'(?<!\w)#([A-Za-z][A-Za-z0-9_]*)'
        else:
            pattern = r'(?<!\w)#([A-Za-z0-9][A-Za-z0-9_]*)'
        hashtags = re.findall(pattern, t)
        unique = list(dict.fromkeys([h.lower() for h in hashtags]))
        return {
            "ok": True,
            "result": unique,
            "hashtags": unique,
            "all_hashtags": [h.lower() for h in hashtags],
            "count": len(unique),
            "platform": platform,
            "has_hashtags": len(hashtags) > 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
