import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        lowercase = bool(event.get("lowercase", False))
        # Tokenize: words, numbers, punctuation
        pattern = r'\b\w+\b|[^\w\s]'
        raw_tokens = re.findall(pattern, text)
        tokens = []
        pos = 0
        for token in raw_tokens:
            start = text.find(token, pos)
            end = start + len(token)
            t = token.lower() if lowercase else token
            tokens.append({
                "token": t,
                "start": start,
                "end": end,
                "is_word": bool(re.match(r'\w+', token)),
                "is_punct": bool(re.match(r'[^\w\s]', token)),
                "is_number": bool(re.match(r'^\d+$', token))
            })
            pos = end
        return {
            "ok": True,
            "result": [t["token"] for t in tokens],
            "tokens": tokens,
            "count": len(tokens),
            "word_count": sum(1 for t in tokens if t["is_word"])
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
