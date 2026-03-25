import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        stripped = re.sub(r'<[^>]+>', ' ', t)
        words = re.findall(r"\b\w+(?:'\w+)?\b", stripped)
        chars = len(t)
        chars_no_spaces = len(t.replace(' ', '').replace('\n', '').replace('\t', ''))
        sentences = len(re.findall(r'[.!?]+', stripped)) or 1
        paragraphs = len([p for p in t.split('\n\n') if p.strip()])
        return {
            "ok": True,
            "result": len(words),
            "word_count": len(words),
            "char_count": chars,
            "char_count_no_spaces": chars_no_spaces,
            "sentence_count": sentences,
            "paragraph_count": paragraphs,
            "unique_words": len(set(w.lower() for w in words))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
