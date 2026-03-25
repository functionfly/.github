import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    keywords = event.get("keywords", [])
    max_highlights = int(event.get("max_highlights", 5))
    context_words = int(event.get("context_words", 15))
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        stripped = re.sub(r'<[^>]+>', ' ', str(text))
        if keywords:
            highlights = []
            for kw in keywords:
                pattern = re.compile(r'\b' + re.escape(str(kw)) + r'\b', re.I)
                for m in pattern.finditer(stripped):
                    start = max(0, m.start() - context_words * 5)
                    end = min(len(stripped), m.end() + context_words * 5)
                    excerpt = stripped[start:end].strip()
                    highlights.append({"keyword": kw, "excerpt": excerpt, "position": m.start()})
                    if len(highlights) >= max_highlights:
                        break
                if len(highlights) >= max_highlights:
                    break
        else:
            sentences = [s.strip() for s in re.split(r'(?<=[.!?])\s+', stripped) if len(s.strip()) > 30]
            highlights = [{"excerpt": s, "position": i} for i, s in enumerate(sentences[:max_highlights])]
        return {"ok": True, "result": highlights, "highlights": highlights, "count": len(highlights)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
