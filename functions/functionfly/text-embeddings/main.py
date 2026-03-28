import hashlib
import math
import re


def _hash_embed(text, dim):
    """Generate a deterministic hash-based embedding vector."""
    words = re.findall(r'\b\w+\b', text.lower())
    vec = [0.0] * dim
    for i, word in enumerate(words):
        h = hashlib.sha256(f"{word}:{i}".encode()).digest()
        for j in range(dim):
            byte_val = h[j % 32]
            vec[j] += (byte_val / 127.5 - 1.0) * (1.0 / (1 + i * 0.1))
    # Normalize
    norm = math.sqrt(sum(x * x for x in vec)) or 1.0
    return [round(x / norm, 6) for x in vec]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        dim = int(event.get("dimensions", 128))
        if dim < 1 or dim > 4096:
            return {"ok": False, "error": "dimensions must be between 1 and 4096"}
        embedding = _hash_embed(str(text), dim)
        return {
            "ok": True,
            "result": embedding,
            "embedding": embedding,
            "dimensions": dim,
            "model": "hash-embed-v1",
            "note": "Deterministic hash-based embedding simulation"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
