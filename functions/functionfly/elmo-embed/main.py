import hashlib
import math
import re


def _elmo_layer_embed(token, position, context_tokens, layer, dim):
    """ELMo-style multi-layer contextual embedding."""
    context_str = " ".join(context_tokens[:5])  # Use nearby context
    h = hashlib.sha256(f"elmo_l{layer}:{token.lower()}:{context_str[:20]}".encode()).digest()
    vec = []
    for i in range(dim):
        idx = i % 32
        if i >= 32:
            h = hashlib.sha256(f"elmo_l{layer}:{token.lower()}:{context_str[:20]}:{i//32}".encode()).digest()
        val = h[idx] / 127.5 - 1.0
        # Add positional bias
        val += math.sin(position * 0.1 + layer * 0.5) * 0.1
        vec.append(val)
    norm = math.sqrt(sum(x * x for x in vec)) or 1.0
    return [round(x / norm, 6) for x in vec]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        dim = int(event.get("dimensions", 1024))
        if dim < 1 or dim > 4096:
            return {"ok": False, "error": "dimensions must be between 1 and 4096"}
        tokens = re.findall(r'\b\w+\b', text)
        if not tokens:
            return {"ok": False, "error": "No tokens found in text"}
        # ELMo has 3 layers: character CNN + 2 LSTM layers
        num_layers = 3
        layer_weights = [0.333, 0.333, 0.334]  # Equal weighting
        token_embeddings = []
        for i, token in enumerate(tokens):
            context = tokens[max(0, i-2):i] + tokens[i+1:min(len(tokens), i+3)]
            layers = []
            for layer in range(num_layers):
                layer_emb = _elmo_layer_embed(token, i, context, layer, dim)
                layers.append(layer_emb)
            # Weighted combination of layers
            combined = [round(sum(layers[l][j] * layer_weights[l] for l in range(num_layers)), 6) for j in range(dim)]
            token_embeddings.append({
                "token": token,
                "position": i,
                "embedding": combined,
                "layers": layers
            })
        # Sentence embedding: mean of token embeddings
        sentence_emb = [round(sum(te["embedding"][j] for te in token_embeddings) / len(token_embeddings), 6) for j in range(dim)]
        return {
            "ok": True,
            "result": sentence_emb,
            "sentence_embedding": sentence_emb,
            "token_embeddings": [{"token": te["token"], "position": te["position"], "embedding": te["embedding"]} for te in token_embeddings],
            "dimensions": dim,
            "num_layers": num_layers,
            "model": "elmo-hash-v1",
            "note": "Deterministic hash-based ELMo simulation — for production use, integrate allennlp or tensorflow-hub"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
