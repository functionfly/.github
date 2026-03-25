import re
from collections import Counter


STOP_WORDS = {"the","a","an","and","or","but","in","on","at","to","for","of","with","by","from","is","are","was","were","be","been","have","has","do","does","will","would","could","should","this","that","it","i","we","you","he","she","they","as","if","not","so","than","too","very","also","such","no","more","most","some","can","may","might","just","then","when","where","who","what","which","how","all","each","both","few","here","there"}


def _score_sentence(sentence, word_freq, total_words):
    words = re.findall(r'\b[a-zA-Z]+\b', sentence.lower())
    if not words:
        return 0
    score = sum(word_freq.get(w, 0) for w in words if w not in STOP_WORDS)
    return score / len(words)


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    max_sentences = int(event.get("max_sentences", 3))
    max_chars = event.get("max_chars")
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        stripped = re.sub(r'<[^>]+>', ' ', str(text))
        sentences = [s.strip() for s in re.split(r'(?<=[.!?])\s+', stripped) if len(s.strip()) > 20]
        if not sentences:
            return {"ok": True, "result": stripped[:500], "summary": stripped[:500], "sentences_used": 1}
        words = re.findall(r'\b[a-zA-Z]+\b', stripped.lower())
        word_freq = Counter(w for w in words if w not in STOP_WORDS)
        scored = [(s, _score_sentence(s, word_freq, len(words)), i) for i, s in enumerate(sentences)]
        top = sorted(scored, key=lambda x: x[1], reverse=True)[:max_sentences]
        top_ordered = sorted(top, key=lambda x: x[2])
        summary = " ".join(t[0] for t in top_ordered)
        if max_chars:
            summary = summary[:int(max_chars)]
        return {"ok": True, "result": summary, "summary": summary, "sentences_used": len(top_ordered), "original_sentences": len(sentences)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
