import re
from collections import defaultdict


def _normalize(text):
    t = re.sub(r'[^\w\s]', '', text.lower())
    return ' '.join(t.split())


def _shingles(text, k=3):
    words = text.split()
    if len(words) < k:
        return set(words)
    return {' '.join(words[i:i+k]) for i in range(len(words) - k + 1)}


def _jaccard(a, b):
    if not a and not b:
        return 1.0
    if not a or not b:
        return 0.0
    intersection = len(a & b)
    union = len(a | b)
    return intersection / union


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    threshold = float(event.get("threshold", 0.8))
    key = event.get("key", "text")
    if not items or not isinstance(items, list):
        return {"ok": False, "error": "items must be a non-empty list"}
    try:
        texts = []
        for item in items:
            if isinstance(item, dict):
                texts.append(_normalize(str(item.get(key, ""))))
            else:
                texts.append(_normalize(str(item)))
        shingles_list = [_shingles(t) for t in texts]
        duplicates = []
        seen = set()
        unique_indices = []
        for i in range(len(items)):
            if i in seen:
                continue
            unique_indices.append(i)
            for j in range(i + 1, len(items)):
                if j in seen:
                    continue
                sim = _jaccard(shingles_list[i], shingles_list[j])
                if sim >= threshold:
                    seen.add(j)
                    duplicates.append({"original_index": i, "duplicate_index": j, "similarity": round(sim, 4)})
        unique_items = [items[i] for i in unique_indices]
        return {
            "ok": True,
            "result": unique_items,
            "unique_items": unique_items,
            "unique_count": len(unique_items),
            "total_count": len(items),
            "duplicates_found": len(duplicates),
            "duplicate_pairs": duplicates,
            "threshold": threshold
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
