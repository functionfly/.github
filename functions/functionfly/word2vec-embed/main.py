import hashlib
import math


def _word2vec_embed(word, dim, seed=42):
    """Deterministic Word2Vec-style embedding using hash."""
    h = hashlib.sha256(f"w2v:{seed}:{word.lower()}".encode()).digest()
    vec = []
    for i in range(dim):
        byte_val = h[i % 32]
        # Use multiple hash rounds for larger dimensions
        if i >= 32:
            h2 = hashlib.sha256(f"w2v:{seed}:{word.lower()}:{i//32}".encode()).digest()
            byte_val = h2[i % 32]
        vec.append(byte_val / 127.5 - 1.0)
    norm = math.sqrt(sum(x * x for x in vec)) or 1.0
    return [round(x / norm, 6) for x in vec]


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    words = event.get("words")
    if not words or not isinstance(words, list):
        return {"ok": False, "error": "words (array of strings) is required"}
    try:
        dim = int(event.get("dimensions", 100))
        if dim < 1 or dim > 1024:
            return {"ok": False, "error": "dimensions must be between 1 and 1024"}
        embeddings = {}
        for word in words:
            if isinstance(word, str):
                embeddings[word] = _word2vec_embed(word, dim)
        return {
            "ok": True,
            "result": embeddings,
            "embeddings": embeddings,
            "dimensions": dim,
            "model": "word2vec-hash-v1",
            "word_count": len(embeddings),
            "note": "Deterministic hash-based Word2Vec simulation — for production use, load pre-trained word2vec vectors"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
