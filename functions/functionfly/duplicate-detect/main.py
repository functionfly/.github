import re
import hashlib


def _normalize(text):
    t = re.sub(r'[^\w\s]', '', text.lower())
    return ' '.join(t.split())


def _fingerprint(text):
    normalized = _normalize(text)
    return hashlib.md5(normalized.encode()).hexdigest()


def _shingles(text, k=4):
    words = text.split()
    if len(words) < k:
        return set(words)
    return {' '.join(words[i:i+k]) for i in range(len(words) - k + 1)}


def handler(event):
    text1 = event.get("text1") if isinstance(event, dict) else None
    text2 = event.get("text2")
    threshold = float(event.get("threshold", 0.8))
    if not text1 or not text2:
        return {"ok": False, "error": "text1 and text2 are required"}
    try:
        t1, t2 = str(text1), str(text2)
        if t1 == t2:
            return {"ok": True, "result": True, "is_duplicate": True, "similarity": 1.0, "method": "exact_match", "threshold": threshold}
        fp1, fp2 = _fingerprint(t1), _fingerprint(t2)
        if fp1 == fp2:
            return {"ok": True, "result": True, "is_duplicate": True, "similarity": 1.0, "method": "normalized_exact", "threshold": threshold}
        n1, n2 = _normalize(t1), _normalize(t2)
        sh1, sh2 = _shingles(n1), _shingles(n2)
        if sh1 and sh2:
            jaccard = len(sh1 & sh2) / len(sh1 | sh2)
        else:
            jaccard = 0.0
        is_dup = jaccard >= threshold
        return {
            "ok": True,
            "result": is_dup,
            "is_duplicate": is_dup,
            "similarity": round(jaccard, 4),
            "threshold": threshold,
            "method": "jaccard_shingles",
            "fingerprint1": fp1,
            "fingerprint2": fp2
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
