import re
from collections import Counter


STOP_WORDS = {"the","a","an","and","or","but","in","on","at","to","for","of","with","by","from","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","need","used","this","that","these","those","it","its","my","your","his","her","our","their","we","you","he","she","they","i","me","us","him","them","what","which","who","when","where","why","how","all","each","every","both","few","more","most","other","some","such","no","not","only","same","so","than","too","very","just","also","as","if","about","after","before","between","into","through","during","above","below","over","under","out","up","down","then","now","here","there","because","while","although","however","therefore"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    top_n = int(event.get("top_n", 10))
    min_length = int(event.get("min_length", 3))
    include_phrases = event.get("include_phrases", True)
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        stripped = re.sub(r'<[^>]+>', ' ', str(text))
        words = re.findall(r'\b[a-zA-Z]+\b', stripped.lower())
        filtered = [w for w in words if len(w) >= min_length and w not in STOP_WORDS]
        word_freq = Counter(filtered)
        keywords = [{"word": w, "count": c, "score": round(c / len(filtered), 4)} for w, c in word_freq.most_common(top_n)]
        phrases = []
        if include_phrases:
            bigrams = [f"{filtered[i]} {filtered[i+1]}" for i in range(len(filtered)-1)]
            phrase_freq = Counter(bigrams)
            phrases = [{"phrase": p, "count": c} for p, c in phrase_freq.most_common(5) if c > 1]
        return {"ok": True, "result": keywords, "keywords": keywords, "phrases": phrases, "total_words": len(words)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
