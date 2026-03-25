import re
import math


def _normalize(text):
    t = re.sub(r'[^\w\s]', '', text.lower())
    return ' '.join(t.split())


def _jaccard(a, b, k=3):
    words_a, words_b = a.split(), b.split()
    def shingles(words):
        if len(words) < k:
            return set(words)
        return {' '.join(words[i:i+k]) for i in range(len(words) - k + 1)}
    sa, sb = shingles(words_a), shingles(words_b)
    if not sa and not sb:
        return 1.0
    if not sa or not sb:
        return 0.0
    return len(sa & sb) / len(sa | sb)


def _cosine(a, b):
    words = list(set(a.split() + b.split()))
    def vec(text):
        counts = {}
        for w in text.split():
            counts[w] = counts.get(w, 0) + 1
        return [counts.get(w, 0) for w in words]
    va, vb = vec(a), vec(b)
    dot = sum(x*y for x, y in zip(va, vb))
    mag_a = math.sqrt(sum(x*x for x in va))
    mag_b = math.sqrt(sum(y*y for y in vb))
    if mag_a == 0 or mag_b == 0:
        return 0.0
    return dot / (mag_a * mag_b)


def _levenshtein(a, b):
    m, n = len(a), len(b)
    dp = list(range(n + 1))
    for i in range(1, m + 1):
        prev = dp[:]
        dp[0] = i
        for j in range(1, n + 1):
            dp[j] = prev[j-1] if a[i-1] == b[j-1] else 1 + min(prev[j], dp[j-1], prev[j-1])
    return 1 - dp[n] / max(m, n) if max(m, n) > 0 else 1.0


def handler(event):
    text1 = event.get("text1") if isinstance(event, dict) else None
    text2 = event.get("text2")
    method = event.get("method", "jaccard")
    if not text1 or not text2:
        return {"ok": False, "error": "text1 and text2 are required"}
    try:
        t1, t2 = _normalize(str(text1)), _normalize(str(text2))
        if method == "cosine":
            score = _cosine(t1, t2)
        elif method == "levenshtein":
            score = _levenshtein(str(text1), str(text2))
        else:
            score = _jaccard(t1, t2)
        score = round(score, 6)
        return {
            "ok": True,
            "result": score,
            "similarity": score,
            "similarity_percent": round(score * 100, 2),
            "method": method,
            "is_similar": score >= 0.7
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
