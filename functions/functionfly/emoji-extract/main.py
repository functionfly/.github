import re


EMOJI_PATTERN = re.compile(
    "["
    "\U0001F600-\U0001F64F"
    "\U0001F300-\U0001F5FF"
    "\U0001F680-\U0001F6FF"
    "\U0001F700-\U0001F77F"
    "\U0001F780-\U0001F7FF"
    "\U0001F800-\U0001F8FF"
    "\U0001F900-\U0001F9FF"
    "\U0001FA00-\U0001FA6F"
    "\U0001FA70-\U0001FAFF"
    "\U00002702-\U000027B0"
    "\U000024C2-\U0001F251"
    "]+",
    flags=re.UNICODE
)


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        emojis = EMOJI_PATTERN.findall(t)
        all_emojis = [c for group in emojis for c in group]
        from collections import Counter
        counts = Counter(all_emojis)
        return {
            "ok": True,
            "result": all_emojis,
            "emojis": all_emojis,
            "unique": list(counts.keys()),
            "counts": dict(counts),
            "count": len(all_emojis),
            "has_emojis": len(all_emojis) > 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
