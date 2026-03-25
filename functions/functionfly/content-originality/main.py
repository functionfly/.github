import re
from collections import Counter


COMMON_PHRASES = {"in conclusion","in summary","it is important","as a result","on the other hand","for example","in addition","furthermore","however","therefore","nevertheless","in contrast","as mentioned","it should be noted","it can be seen","in order to","due to the fact","at the end of the day","needless to say","last but not least","in terms of","with regards to","in the context of","the fact that"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text).lower()
        words = re.findall(r'\b[a-zA-Z]+\b', t)
        n = len(words)
        if n == 0:
            return {"ok": False, "error": "No words found"}
        unique_ratio = len(set(words)) / n
        bigrams = [f"{words[i]} {words[i+1]}" for i in range(n-1)]
        bigram_freq = Counter(bigrams)
        repetition_score = sum(v for v in bigram_freq.values() if v > 2) / max(len(bigrams), 1)
        cliche_count = sum(1 for phrase in COMMON_PHRASES if phrase in t)
        cliche_penalty = min(0.3, cliche_count * 0.05)
        avg_word_len = sum(len(w) for w in words) / n
        vocab_score = min(1.0, avg_word_len / 6 * 0.3 + unique_ratio * 0.7)
        originality = round(max(0, min(1, vocab_score - repetition_score - cliche_penalty)), 4)
        return {
            "ok": True,
            "result": originality,
            "originality_score": originality,
            "unique_word_ratio": round(unique_ratio, 4),
            "repetition_score": round(repetition_score, 4),
            "cliche_count": cliche_count,
            "grade": "highly original" if originality > 0.7 else ("original" if originality > 0.5 else ("average" if originality > 0.3 else "low originality"))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
