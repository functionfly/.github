import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        urls = re.findall(r'https?://[^\s<>"\']+', t)
        emails = re.findall(r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b', t)
        phones = re.findall(r'(?:\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}', t)
        dates = re.findall(r'\b(?:\d{1,2}[-/]\d{1,2}[-/]\d{2,4}|\d{4}[-/]\d{1,2}[-/]\d{1,2}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{1,2},?\s+\d{4})\b', t, re.I)
        hashtags = re.findall(r'#([A-Za-z0-9_]+)', t)
        mentions = re.findall(r'@([A-Za-z0-9_]{1,50})', t)
        cashtags = re.findall(r'\$([A-Z]{1,6})\b', t)
        capitalized = list(set(re.findall(r'\b[A-Z][a-z]{2,}(?:\s+[A-Z][a-z]+)*\b', t)))
        return {
            "ok": True,
            "result": {"urls": urls, "emails": emails, "phones": phones},
            "entities": {
                "urls": list(set(urls)),
                "emails": list(set(emails)),
                "phones": list(set(phones)),
                "dates": list(set(dates)),
                "hashtags": list(set(hashtags)),
                "mentions": list(set(mentions)),
                "cashtags": list(set(cashtags)),
                "proper_nouns": capitalized[:20],
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
