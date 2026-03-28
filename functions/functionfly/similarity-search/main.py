import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","have","has","had","do","does","did","will","would","could","should","to","of","in","for","on","with","at","by","from","and","but","or","not","this","that","it","its","i","you","he","she","we","they"}


def _tokenize(text):
    return [w for w in re.findall(r'\b[a-z]+\b', text.lower()) if w not in STOPWORDS]


def _tfidf_vec(tokens, all_docs):
    tf = Counter(tokens)
    n = len(tokens) or 1
    N = len(all_docs)
    vec = {}
    for term, count in tf.items():
        df = sum(1 for doc in all_docs if term in doc)
        idf = math.log((N + 1) / (df + 1)) + 1
        vec[term] = (count / n) * idf
    return vec


def _cosine(v1, v2):
    keys = set(v1) & set(v2)
    dot = sum(v1[k] * v2[k] for k in keys)
    mag1 = math.sqrt(sum(x * x for x in v1.values())) or 1e-9
    mag2 = math.sqrt(sum(x * x for x in v2.values())) or 1e-9
    return dot / (mag1 * mag2)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    query = event.get("query")
    candidates = event.get("candidates")
    if not query or not isinstance(query, str):
        return {"ok": False, "error": "query (string) is required"}
    if not candidates or not isinstance(candidates, list):
        return {"ok": False, "error": "candidates (array of strings) is required"}
    try:
        top_k = int(event.get("top_k", 5))
        all_texts = [query] + [str(c) for c in candidates]
        all_tokens = [_tokenize(t) for t in all_texts]
        query_vec = _tfidf_vec(all_tokens[0], all_tokens)
        results = []
        for i, (cand, tokens) in enumerate(zip(candidates, all_tokens[1:])):
            cand_vec = _tfidf_vec(tokens, all_tokens)
            score = round(_cosine(query_vec, cand_vec), 4)
            results.append({"text": str(cand), "score": score, "index": i})
        results.sort(key=lambda x: x["score"], reverse=True)
        matches = results[:top_k]
        return {
            "ok": True,
            "result": matches,
            "matches": matches,
            "total_candidates": len(candidates)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
