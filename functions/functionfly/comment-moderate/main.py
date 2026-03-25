BANNED_WORDS = ["spam","scam","phishing","hack","malware","virus","bot","fake","abuse","harassment","bully","slur","profane","threats","advertisement","promo","discount","buy now","click here"]
SPAM_URLS = re.compile(r'https?://[^\s]+') if False else None


import re


TOXIC = {"fuck","shit","bitch","cunt","asshole","bastard","damn","hate","idiot","stupid","moron","loser","ugly","fat","dumb"}
SPAM_KW = ["buy now","click here","limited offer","free money","earn money","work from home","lose weight fast","enlarge","pill","casino","xxx","adult"]
SPAM_URL_PATTERN = re.compile(r'https?://', re.I)


def handler(event):
    comment = event.get("comment") if isinstance(event, dict) else None
    rules = event.get("rules", {})
    if not comment:
        return {"ok": False, "error": "comment is required"}
    try:
        text = str(comment).lower()
        words = text.split()
        issues = []
        action = "approve"
        toxic_words = [w.strip(".,!?") for w in words if w.strip(".,!?") in TOXIC]
        if toxic_words:
            issues.append({"type": "toxicity", "details": toxic_words[:3]})
            action = "flag"
        spam_kw = [kw for kw in SPAM_KW if kw in text]
        if spam_kw:
            issues.append({"type": "spam", "details": spam_kw[:3]})
            action = "reject"
        url_count = len(SPAM_URL_PATTERN.findall(comment))
        max_urls = int(rules.get("max_urls", 2))
        if url_count > max_urls:
            issues.append({"type": "excessive_urls", "count": url_count})
            action = "flag"
        max_caps = float(rules.get("max_caps_ratio", 0.5))
        caps_ratio = sum(1 for c in comment if c.isupper()) / max(len(comment), 1)
        if caps_ratio > max_caps and len(comment) > 20:
            issues.append({"type": "excessive_caps", "ratio": round(caps_ratio, 3)})
            if action == "approve": action = "flag"
        min_length = int(rules.get("min_length", 2))
        if len(comment.strip()) < min_length:
            issues.append({"type": "too_short"})
            action = "flag"
        return {"ok": True, "result": action, "action": action, "issues": issues, "issue_count": len(issues), "approved": action == "approve"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
