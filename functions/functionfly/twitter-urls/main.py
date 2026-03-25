import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        urls = re.findall(r'https?://[^\s<>"\']+', str(text))
        unique = list(dict.fromkeys(urls))
        return {"ok": True, "result": unique, "urls": unique, "count": len(unique)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
