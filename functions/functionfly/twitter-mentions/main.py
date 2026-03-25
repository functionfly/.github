import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        mentions = re.findall(r'@([A-Za-z0-9_]{1,50})', str(text))
        unique = list(dict.fromkeys(m.lower() for m in mentions))
        return {"ok": True, "result": unique, "mentions": unique, "count": len(unique), "all": [m.lower() for m in mentions]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
