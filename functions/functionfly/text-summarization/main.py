import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","to","of","in","for","on","with","at","by","from","as","into","and","but","or","not","this","that","it","its","i","you","he","she","we","they","me","him","her","us","them","my","your","his","our","their","what","which","who","when","where","why","how"}


def _tokenize(text):
    return [w for w in re.findall(r'\b[a-z]+\b', text.lower()) if w not in STOPWORDS and len(w) > 2]


def _split_sentences(text):
    sents = re.split(r'(?<=[.!?])\s+', text.strip())
    return [s.strip() for s in sents if len(s.strip()) > 10]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        sentences = _split_sentences(text)
        if not sentences:
            return {"ok": False, "error": "No sentences found in text"}
        num_sentences = int(event.get("num_sentences", 3))
        ratio = float(event.get("ratio", 0.3))
        target = max(1, min(num_sentences, max(1, int(len(sentences) * ratio))))
        target = min(target, len(sentences))

        # Build TF-IDF
        all_tokens = [_tokenize(s) for s in sentences]
        N = len(sentences)
        # IDF
        df = Counter()
        for tokens in all_tokens:
            for t in set(tokens):
                df[t] += 1
        idf = {t: math.log((N + 1) / (c + 1)) + 1 for t, c in df.items()}

        # Score sentences
        scored = []
        for i, (sent, tokens) in enumerate(zip(sentences, all_tokens)):
            if not tokens:
                scored.append((0.0, i, sent))
                continue
            tf = Counter(tokens)
            n = len(tokens)
            score = sum((tf[t] / n) * idf.get(t, 1.0) for t in tf)
            # Bonus for position (first and last sentences)
            if i == 0:
                score *= 1.5
            elif i == len(sentences) - 1:
                score *= 1.2
            scored.append((score, i, sent))

        # Select top sentences, preserve order
        top = sorted(scored, key=lambda x: x[0], reverse=True)[:target]
        top_ordered = sorted(top, key=lambda x: x[1])
        summary_sentences = [s for _, _, s in top_ordered]
        summary = " ".join(summary_sentences)

        return {
            "ok": True,
            "result": summary,
            "summary": summary,
            "sentences": summary_sentences,
            "original_sentences": len(sentences),
            "summary_sentences": len(summary_sentences),
            "compression_ratio": round(len(summary_sentences) / len(sentences), 4)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
