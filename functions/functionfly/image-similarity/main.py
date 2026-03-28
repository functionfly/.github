import hashlib
import math

def _phash(seed):
    """Perceptual hash simulation."""
    h = hashlib.sha256(seed.encode()).digest()
    return [1 if b > 127 else 0 for b in h[:64]]

def _hamming_similarity(h1, h2):
    matches = sum(a == b for a, b in zip(h1, h2))
    return round(matches / len(h1), 4)

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    url1 = event.get("image_url_1", "")
    url2 = event.get("image_url_2", "")
    b64_1 = event.get("image_base64_1", "")
    b64_2 = event.get("image_base64_2", "")
    seed1 = url1 or b64_1[:50]
    seed2 = url2 or b64_2[:50]
    if not seed1 or not seed2:
        return {"ok": False, "error": "Two images are required (image_url_1/2 or image_base64_1/2)"}
    try:
        hash1 = _phash(seed1)
        hash2 = _phash(seed2)
        similarity = _hamming_similarity(hash1, hash2)
        # Also compute embedding similarity
        emb1 = [round((hashlib.sha256(f"emb:{seed1}:{i}".encode()).digest()[0] / 127.5 - 1.0), 4) for i in range(32)]
        emb2 = [round((hashlib.sha256(f"emb:{seed2}:{i}".encode()).digest()[0] / 127.5 - 1.0), 4) for i in range(32)]
        dot = sum(a * b for a, b in zip(emb1, emb2))
        mag1 = math.sqrt(sum(x*x for x in emb1)) or 1e-9
        mag2 = math.sqrt(sum(x*x for x in emb2)) or 1e-9
        cosine_sim = round(dot / (mag1 * mag2), 4)
        combined = round((similarity * 0.4 + (cosine_sim + 1) / 2 * 0.6), 4)
        return {
            "ok": True,
            "result": combined,
            "similarity": combined,
            "phash_similarity": similarity,
            "embedding_similarity": cosine_sim,
            "is_similar": combined > 0.7,
            "model": "mock-phash-v1",
            "note": "Mock image similarity — for production use, integrate a vision model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
