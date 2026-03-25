import re


SPAM_PATTERNS = ["click here","buy now","limited time","act now","free offer","guaranteed","no risk","earn money","make money","work from home","extra income","lose weight","lowest price","order now","special offer","congratulations","winner","prize","claim","urgent","100%","money back","no obligation","cancel anytime","opt in","unsubscribe","dear friend","you have been selected","exclusively for you","as seen on","casino","poker","pills","medication","drugs","enhancement","replica","enlarge","xxx","adult content","call now","toll free"]
SPAM_INDICATORS = ["!!!","$$$","FREE","WIN","WINNER","GUARANTEED","URGENT","CLICK NOW","LIMITED TIME"]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    subject = event.get("subject", "")
    if not text and not subject:
        return {"ok": False, "error": "text is required"}
    try:
        combined = (str(subject) + " " + str(text or "")).strip()
        lower = combined.lower()
        pattern_matches = [p for p in SPAM_PATTERNS if p in lower]
        indicator_matches = [ind for ind in SPAM_INDICATORS if ind in combined]
        caps_ratio = sum(1 for c in combined if c.isupper()) / max(len(combined), 1)
        exclamation_count = combined.count('!')
        url_count = len(re.findall(r'https?://', combined))
        spam_score = min(1.0, len(pattern_matches) * 0.1 + len(indicator_matches) * 0.15 + (caps_ratio > 0.3) * 0.2 + min(exclamation_count * 0.05, 0.3) + min(url_count * 0.05, 0.2))
        is_spam = spam_score >= 0.4
        return {
            "ok": True,
            "result": spam_score,
            "spam_score": round(spam_score, 4),
            "is_spam": is_spam,
            "confidence": "high" if spam_score > 0.7 else ("medium" if spam_score > 0.4 else "low"),
            "spam_indicators": pattern_matches[:5] + indicator_matches[:3],
            "caps_ratio": round(caps_ratio, 4)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
