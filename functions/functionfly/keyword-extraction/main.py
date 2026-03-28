import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","to","of","in","for","on","with","at","by","from","as","into","and","but","or","nor","not","this","that","these","those","it","its","i","you","he","she","we","they","me","him","her","us","them","my","your","his","our","their","what","which","who","when","where","why","how","all","each","every","both","few","more","most","other","some","such","no","only","same","so","than","too","very","just","also","about","up","out","if","then","than","because","while","although","though","since","unless","until","after","before","above","below","between","through","during","over","under","again","further","once"}


def _ngrams(tokens, n):
    return [" ".join(tokens[i:i+n]) for i in range(len(tokens) - n + 1)]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        top_n = int(event.get("top_n", 10))
        ngram_max = min(int(event.get("ngram_max", 2)), 4)
        tokens = [w for w in re.findall(r'\b[a-z]+\b', text.lower()) if w not in STOPWORDS and len(w) > 2]
        if not tokens:
            return {"ok": True, "result": [], "keywords": []}
        # Build candidate phrases
        candidates = []
        for n in range(1, ngram_max + 1):
            candidates.extend(_ngrams(tokens, n))
        freq = Counter(candidates)
        # Score: frequency * avg word length bonus
        scored = []
        for phrase, count in freq.items():
            words = phrase.split()
            avg_len = sum(len(w) for w in words) / len(words)
            score = round(count * (1 + avg_len / 10), 4)
            scored.append({"keyword": phrase, "score": score, "count": count})
        scored.sort(key=lambda x: x["score"], reverse=True)
        # Deduplicate: skip phrases whose words are all in a higher-ranked phrase
        seen_words = set()
        keywords = []
        for item in scored:
            words = set(item["keyword"].split())
            if not words.issubset(seen_words):
                keywords.append(item)
                seen_words.update(words)
            if len(keywords) >= top_n:
                break
        return {
            "ok": True,
            "result": keywords,
            "keywords": keywords
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
