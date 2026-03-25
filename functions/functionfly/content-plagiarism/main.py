import re


def _shingles(text, k=5):
    words = re.findall(r'\b[a-z]+\b', text.lower())
    return set(' '.join(words[i:i+k]) for i in range(max(len(words)-k+1, 1)))


def handler(event):
    text_a = event.get("text_a") if isinstance(event, dict) else None
    text_b = event.get("text_b")
    threshold = float(event.get("threshold", 0.5))
    shingle_size = int(event.get("shingle_size", 5))
    if not text_a or not text_b:
        return {"ok": False, "error": "text_a and text_b are required"}
    try:
        shingles_a = _shingles(str(text_a), shingle_size)
        shingles_b = _shingles(str(text_b), shingle_size)
        if not shingles_a or not shingles_b:
            return {"ok": True, "result": 0, "similarity": 0, "is_plagiarized": False}
        intersection = len(shingles_a & shingles_b)
        union = len(shingles_a | shingles_b)
        jaccard = round(intersection / union, 4) if union else 0
        is_plagiarized = jaccard >= threshold
        return {
            "ok": True,
            "result": jaccard,
            "similarity": jaccard,
            "is_plagiarized": is_plagiarized,
            "threshold": threshold,
            "shared_shingles": intersection,
            "total_shingles_a": len(shingles_a),
            "total_shingles_b": len(shingles_b)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
