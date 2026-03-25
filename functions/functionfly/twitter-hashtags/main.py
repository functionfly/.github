import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        hashtags = re.findall(r'#([A-Za-z0-9_]+)', str(text))
        unique = list(dict.fromkeys(h.lower() for h in hashtags))
        return {"ok": True, "result": unique, "hashtags": unique, "count": len(unique)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
