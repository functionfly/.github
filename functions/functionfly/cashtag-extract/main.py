import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        # Cashtags: $ followed by 1-6 uppercase letters (stock tickers)
        cashtags = re.findall(r'(?<!\w)\$([A-Z]{1,6})(?!\w)', t)
        # Also catch lowercase and normalize
        all_tags = re.findall(r'(?<!\w)\$([A-Za-z]{1,6})(?!\w)', t)
        normalized = list(dict.fromkeys([tag.upper() for tag in all_tags]))
        return {
            "ok": True,
            "result": normalized,
            "cashtags": normalized,
            "count": len(normalized),
            "has_cashtags": len(normalized) > 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
