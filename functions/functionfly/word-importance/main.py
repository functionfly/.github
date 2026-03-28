import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","to","of","in","for","on","with","at","by","from","as","into","and","but","or","not","this","that","it","its","i","you","he","she","we","they","me","him","her","us","them","my","your","his","our","their","what","which","who","when","where","why","how","all","each","every","both","few","more","most","other","some","such","no","only","same","so","than","too","very","just","also","about","up","out","if","then","than","because","while","although","though","since","unless","until","after","before","above","below","between","through","during","over","under","again","further","once"}


def _tokenize(text):
    return [w for w in re.findall(r'\b[a-z]+\b', text.lower()) if w not in STOPWORDS and len(w) > 2]


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    documents = event.get("documents")
    if not documents or not isinstance(documents, list):
        return {"ok": False, "error": "documents (array of strings) is required"}
    try:
        query = event.get("query")
        docs_tokens = [_tokenize(str(doc)) for doc in documents]
        N = len(docs_tokens)
        # Compute IDF
        df = Counter()
        for tokens in docs_tokens:
            for t in set(tokens):
                df[t] += 1
        idf = {t: math.log((N + 1) / (c + 1)) + 1 for t, c in df.items()}
        # Compute TF-IDF for each document
        results = []
        for i, (doc, tokens) in enumerate(zip(documents, docs_tokens)):
            if not tokens:
                results.append({"document_index": i, "words": []})
                continue
            tf = Counter(tokens)
            n = len(tokens)
            word_scores = []
            for term, count in tf.items():
                tfidf = (count / n) * idf.get(term, 1.0)
                word_scores.append({"word": term, "tfidf": round(tfidf, 6), "tf": round(count / n, 6), "idf": round(idf.get(term, 1.0), 6), "count": count})
            word_scores.sort(key=lambda x: x["tfidf"], reverse=True)
            results.append({"document_index": i, "words": word_scores[:20]})
        # If query provided, score query terms across all docs
        query_scores = None
        if query:
            q_tokens = _tokenize(str(query))
            query_scores = []
            for term in q_tokens:
                doc_scores = []
                for i, tokens in enumerate(docs_tokens):
                    tf = Counter(tokens)
                    n = len(tokens) or 1
                    tfidf = (tf.get(term, 0) / n) * idf.get(term, 1.0)
                    doc_scores.append({"document_index": i, "tfidf": round(tfidf, 6)})
                query_scores.append({"term": term, "idf": round(idf.get(term, 0.0), 6), "document_scores": doc_scores})
        return {
            "ok": True,
            "result": results,
            "documents": results,
            "query_scores": query_scores,
            "vocabulary_size": len(df)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
