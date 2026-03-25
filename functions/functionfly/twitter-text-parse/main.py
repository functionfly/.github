import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        mentions = list(set(re.findall(r'@([A-Za-z0-9_]+)', t)))
        hashtags = list(set(re.findall(r'#([A-Za-z0-9_]+)', t)))
        cashtags = list(set(re.findall(r'\$([A-Z]{1,6})', t)))
        urls = re.findall(r'https?://[^\s]+', t)
        # Twitter character count: URLs = 23, CJK = 2
        char_count = len(t)
        url_placeholder_len = 23 * len(urls) - sum(len(u) for u in urls)
        twitter_count = char_count + url_placeholder_len
        return {
            "ok": True,
            "result": t,
            "mentions": mentions,
            "hashtags": hashtags,
            "cashtags": cashtags,
            "urls": urls,
            "char_count": twitter_count,
            "over_limit": twitter_count > 280
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
