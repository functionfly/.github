import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    platform = event.get("platform", "twitter")
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        if platform == "twitter":
            pattern = r'(?<!\w)@([A-Za-z0-9_]{1,15})'
        elif platform == "instagram":
            pattern = r'(?<!\w)@([A-Za-z0-9_.]{1,30})'
        elif platform == "github":
            pattern = r'(?<!\w)@([A-Za-z0-9-]{1,39})'
        else:
            pattern = r'(?<!\w)@([A-Za-z0-9_.]{1,50})'
        mentions = re.findall(pattern, t)
        unique_mentions = list(dict.fromkeys(mentions))
        return {
            "ok": True,
            "result": unique_mentions,
            "mentions": unique_mentions,
            "all_mentions": mentions,
            "count": len(unique_mentions),
            "platform": platform,
            "has_mentions": len(mentions) > 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
