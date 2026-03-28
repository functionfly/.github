import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","need","dare","ought","used","to","of","in","for","on","with","at","by","from","as","into","through","during","before","after","above","below","between","out","off","over","under","again","further","then","once","and","but","or","nor","so","yet","both","either","neither","not","only","own","same","than","too","very","just","because","if","while","although","though","since","unless","until","when","where","who","which","that","this","these","those","i","you","he","she","it","we","they","me","him","her","us","them","my","your","his","its","our","their"}


def _tokenize(text):
    return [w for w in re.findall(r'\b[a-z]+\b', text.lower()) if w not in STOPWORDS]


def _cosine(v1, v2):
    all_keys = set(v1) | set(v2)
    dot = sum(v1.get(k, 0) * v2.get(k, 0) for k in all_keys)
    mag1 = math.sqrt(sum(x * x for x in v1.values())) or 1e-9
    mag2 = math.sqrt(sum(x * x for x in v2.values())) or 1e-9
    return dot / (mag1 * mag2)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    s1 = event.get("sentence1")
    s2 = event.get("sentence2")
    if not s1 or not isinstance(s1, str):
        return {"ok": False, "error": "sentence1 (string) is required"}
    if not s2 or not isinstance(s2, str):
        return {"ok": False, "error": "sentence2 (string) is required"}
    try:
        t1 = _tokenize(s1)
        t2 = _tokenize(s2)
        if not t1 or not t2:
            return {"ok": True, "result": 0.0, "similarity": 0.0, "jaccard": 0.0, "cosine": 0.0}
        c1 = Counter(t1)
        c2 = Counter(t2)
        cosine = round(_cosine(c1, c2), 4)
        set1, set2 = set(t1), set(t2)
        inter = len(set1 & set2)
        union = len(set1 | set2)
        jaccard = round(inter / union if union else 0.0, 4)
        similarity = round((cosine * 0.7 + jaccard * 0.3), 4)
        return {
            "ok": True,
            "result": similarity,
            "similarity": similarity,
            "cosine": cosine,
            "jaccard": jaccard
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
