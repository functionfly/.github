import hashlib
import math
import re


def _positional_encoding(position, dim):
    """Sinusoidal positional encoding."""
    pe = []
    for i in range(dim):
        if i % 2 == 0:
            pe.append(math.sin(position / (10000 ** (i / dim))))
        else:
            pe.append(math.cos(position / (10000 ** ((i - 1) / dim))))
    return pe


def _attention_embed(token, position, all_tokens, dim, num_heads=8):
    """Transformer-style multi-head attention embedding."""
    # Token embedding
    h = hashlib.sha256(f"transformer:{token.lower()}".encode()).digest()
    token_vec = []
    for i in range(dim):
        idx = i % 32
        if i >= 32:
            h = hashlib.sha256(f"transformer:{token.lower()}:{i//32}".encode()).digest()
        token_vec.append(h[idx] / 127.5 - 1.0)
    # Add positional encoding
    pe = _positional_encoding(position, dim)
    combined = [token_vec[i] + pe[i] * 0.1 for i in range(dim)]
    # Simulate multi-head attention (simplified)
    head_size = dim // num_heads
    attended = list(combined)
    for head in range(num_heads):
        start = head * head_size
        end = start + head_size
        # Attention to context tokens
        for j, ctx_token in enumerate(all_tokens[:8]):  # Limit context
            if ctx_token == token:
                continue
            ctx_h = hashlib.sha256(f"transformer_attn:{head}:{ctx_token.lower()}:{token.lower()}".encode()).digest()
            attn_weight = (ctx_h[0] / 255.0) * 0.1  # Small attention weight
            ctx_vec_h = hashlib.sha256(f"transformer:{ctx_token.lower()}".encode()).digest()
            for i in range(start, min(end, dim)):
                attended[i] += attn_weight * (ctx_vec_h[i % 32] / 127.5 - 1.0)
    norm = math.sqrt(sum(x * x for x in attended)) or 1.0
    return [round(x / norm, 6) for x in attended]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        dim = int(event.get("dimensions", 512))
        if dim < 8 or dim > 4096:
            return {"ok": False, "error": "dimensions must be between 8 and 4096"}
        model = event.get("model", "base")
        num_heads = 12 if model == "large" else 8
        tokens = re.findall(r'\b\w+\b|[^\w\s]', text)
        if not tokens:
            return {"ok": False, "error": "No tokens found in text"}
        token_embeddings = []
        for i, token in enumerate(tokens):
            emb = _attention_embed(token, i, tokens, dim, num_heads)
            token_embeddings.append({"token": token, "position": i, "embedding": emb})
        # Mean pooling for sentence embedding
        sentence_emb = [round(sum(te["embedding"][j] for te in token_embeddings) / len(token_embeddings), 6) for j in range(dim)]
        return {
            "ok": True,
            "result": sentence_emb,
            "sentence_embedding": sentence_emb,
            "token_embeddings": token_embeddings,
            "dimensions": dim,
            "num_heads": num_heads,
            "model": f"transformer-{model}-hash-v1",
            "note": "Deterministic hash-based transformer simulation — for production use, integrate sentence-transformers"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
