import hashlib
import math


def _glove_embed(word, dim, seed=137):
    """Deterministic GloVe-style embedding using hash with co-occurrence simulation."""
    # GloVe uses global co-occurrence statistics; we simulate with different hash seeds
    h1 = hashlib.sha256(f"glove_w:{seed}:{word.lower()}".encode()).digest()
    h2 = hashlib.sha256(f"glove_c:{seed}:{word.lower()}".encode()).digest()
    vec = []
    for i in range(dim):
        idx = i % 32
        if i >= 32:
            h1 = hashlib.sha256(f"glove_w:{seed}:{word.lower()}:{i//32}".encode()).digest()
            h2 = hashlib.sha256(f"glove_c:{seed}:{word.lower()}:{i//32}".encode()).digest()
        # Combine word and context vectors (GloVe averages them)
        w_val = h1[idx] / 127.5 - 1.0
        c_val = h2[idx] / 127.5 - 1.0
        vec.append((w_val + c_val) / 2.0)
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
                embeddings[word] = _glove_embed(word, dim)
        return {
            "ok": True,
            "result": embeddings,
            "embeddings": embeddings,
            "dimensions": dim,
            "model": "glove-hash-v1",
            "word_count": len(embeddings),
            "note": "Deterministic hash-based GloVe simulation — for production use, load pre-trained GloVe vectors"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
