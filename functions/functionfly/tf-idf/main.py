import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","have","has","had","do","does","did","will","would","could","should","to","of","in","for","on","with","at","by","from","and","but","or","not","this","that","it","its","i","you","he","she","we","they","me","him","her","us","them","my","your","his","our","their"}


def _tokenize(text, remove_stopwords=True):
    tokens = re.findall(r'\b[a-z]+\b', text.lower())
    if remove_stopwords:
        return [t for t in tokens if t not in STOPWORDS and len(t) > 1]
    return tokens


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    documents = event.get("documents")
    if not documents or not isinstance(documents, list):
        return {"ok": False, "error": "documents (array of strings) is required"}
    try:
        query = event.get("query")
        remove_stopwords = bool(event.get("remove_stopwords", True))
        docs_tokens = [_tokenize(str(doc), remove_stopwords) for doc in documents]
        N = len(docs_tokens)
        # Compute document frequency
        df = Counter()
        for tokens in docs_tokens:
            for t in set(tokens):
                df[t] += 1
        # Compute IDF
        idf = {t: round(math.log((N + 1) / (c + 1)) + 1, 6) for t, c in df.items()}
        # Compute TF-IDF matrix
        tfidf_matrix = []
        for i, tokens in enumerate(docs_tokens):
            if not tokens:
                tfidf_matrix.append({})
                continue
            tf = Counter(tokens)
            n = len(tokens)
            doc_tfidf = {t: round((count / n) * idf.get(t, 1.0), 6) for t, count in tf.items()}
            tfidf_matrix.append(doc_tfidf)
        # Format results
        doc_results = []
        for i, (doc, tfidf) in enumerate(zip(documents, tfidf_matrix)):
            top_terms = sorted(tfidf.items(), key=lambda x: x[1], reverse=True)[:15]
            doc_results.append({
                "document_index": i,
                "document_preview": str(doc)[:100],
                "top_terms": [{"term": t, "tfidf": s} for t, s in top_terms]
            })
        # Query scoring
        query_result = None
        if query:
            q_tokens = _tokenize(str(query), remove_stopwords)
            query_scores = []
            for i, tfidf in enumerate(tfidf_matrix):
                score = sum(tfidf.get(t, 0) for t in q_tokens)
                query_scores.append({"document_index": i, "score": round(score, 6)})
            query_scores.sort(key=lambda x: x["score"], reverse=True)
            query_result = {"query": query, "document_scores": query_scores}
        return {
            "ok": True,
            "result": doc_results,
            "documents": doc_results,
            "query_result": query_result,
            "vocabulary_size": len(df),
            "idf": {k: v for k, v in sorted(idf.items(), key=lambda x: x[1], reverse=True)[:20]}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
