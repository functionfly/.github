import re
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","have","has","had","do","does","did","will","would","could","should","to","of","in","for","on","with","at","by","from","and","but","or","not","this","that","it","its","i","you","he","she","we","they"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        vocabulary = event.get("vocabulary")
        remove_stopwords = bool(event.get("remove_stopwords", False))
        binary = bool(event.get("binary", False))
        tokens = re.findall(r'\b[a-z]+\b', text.lower())
        if remove_stopwords:
            tokens = [t for t in tokens if t not in STOPWORDS]
        freq = Counter(tokens)
        if vocabulary:
            # Use provided vocabulary
            vocab = [str(v).lower() for v in vocabulary]
            if binary:
                bow = {v: 1 if v in freq else 0 for v in vocab}
            else:
                bow = {v: freq.get(v, 0) for v in vocab}
        else:
            # Build vocabulary from text
            vocab = sorted(freq.keys())
            if binary:
                bow = {v: 1 for v in vocab}
            else:
                bow = dict(freq)
        # Create dense vector
        vector = [bow.get(v, 0) for v in vocab]
        return {
            "ok": True,
            "result": bow,
            "bow": bow,
            "vector": vector,
            "vocabulary": vocab,
            "vocabulary_size": len(vocab),
            "total_tokens": len(tokens),
            "non_zero": sum(1 for v in vector if v > 0)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
