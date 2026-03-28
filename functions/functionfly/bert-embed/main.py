import hashlib
import math
import re


def _bert_token_embed(token, position, context_hash, dim):
    """BERT-style contextual embedding with positional encoding."""
    # Token embedding
    h = hashlib.sha256(f"bert_tok:{token.lower()}".encode()).digest()
    # Positional encoding (sinusoidal)
    pos_enc = []
    for i in range(dim):
        if i % 2 == 0:
            pos_enc.append(math.sin(position / (10000 ** (i / dim))))
        else:
            pos_enc.append(math.cos(position / (10000 ** ((i - 1) / dim))))
    # Context influence
    ctx_h = hashlib.sha256(f"bert_ctx:{context_hash}:{token.lower()}".encode()).digest()
    vec = []
    for i in range(dim):
        idx = i % 32
        if i >= 32:
            h = hashlib.sha256(f"bert_tok:{token.lower()}:{i//32}".encode()).digest()
        tok_val = h[idx] / 127.5 - 1.0
        ctx_val = ctx_h[idx % 32] / 127.5 - 1.0
        combined = tok_val * 0.6 + ctx_val * 0.3 + pos_enc[i] * 0.1
        vec.append(combined)
    norm = math.sqrt(sum(x * x for x in vec)) or 1.0
    return [round(x / norm, 6) for x in vec]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        dim = int(event.get("dimensions", 768))
        if dim < 1 or dim > 4096:
            return {"ok": False, "error": "dimensions must be between 1 and 4096"}
        tokens = re.findall(r'\b\w+\b|[^\w\s]', text)
        # Add [CLS] and [SEP] tokens
        all_tokens = ["[CLS]"] + tokens + ["[SEP]"]
        context_hash = hashlib.md5(text.encode()).hexdigest()[:8]
        token_embeddings = []
        for i, token in enumerate(all_tokens):
            emb = _bert_token_embed(token, i, context_hash, dim)
            token_embeddings.append({"token": token, "position": i, "embedding": emb})
        # CLS token embedding as sentence representation
        cls_embedding = token_embeddings[0]["embedding"]
        # Mean pooling
        mean_pool = [round(sum(te["embedding"][j] for te in token_embeddings) / len(token_embeddings), 6) for j in range(dim)]
        return {
            "ok": True,
            "result": cls_embedding,
            "cls_embedding": cls_embedding,
            "mean_pooled": mean_pool,
            "token_embeddings": token_embeddings,
            "tokens": all_tokens,
            "dimensions": dim,
            "model": "bert-hash-v1",
            "note": "Deterministic hash-based BERT simulation — for production use, integrate transformers library"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
