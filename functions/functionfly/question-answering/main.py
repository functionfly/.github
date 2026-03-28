import re

STOPWORDS = {"a","an","the","is","are","was","were","be","been","have","has","had","do","does","did","will","would","could","should","to","of","in","for","on","with","at","by","from","and","but","or","not","this","that","it","its","i","you","he","she","we","they","me","him","her","us","them","what","who","where","when","why","how","which","whose","whom"}

QUESTION_WORDS = {"who": "PERSON", "what": "THING", "where": "LOCATION", "when": "TIME", "why": "REASON", "how": "METHOD", "which": "CHOICE", "whose": "PERSON", "whom": "PERSON"}


def _tokenize(text):
    return [w.lower() for w in re.findall(r'\b\w+\b', text)]


def _split_sentences(text):
    return re.split(r'(?<=[.!?])\s+', text.strip())


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    question = event.get("question")
    context = event.get("context")
    if not question or not isinstance(question, str):
        return {"ok": False, "error": "question (string) is required"}
    if not context or not isinstance(context, str):
        return {"ok": False, "error": "context (string) is required"}
    try:
        q_tokens = set(_tokenize(question)) - STOPWORDS
        sentences = _split_sentences(context)
        if not sentences:
            return {"ok": False, "error": "No sentences found in context"}

        # Score each sentence by keyword overlap with question
        best_score = -1
        best_sent = ""
        best_idx = 0
        for i, sent in enumerate(sentences):
            s_tokens = set(_tokenize(sent)) - STOPWORDS
            overlap = len(q_tokens & s_tokens)
            # Bonus for question-type words
            q_lower = question.lower()
            for qw in QUESTION_WORDS:
                if qw in q_lower and any(w in sent.lower() for w in ["is","was","are","were","has","have","had"]):
                    overlap += 0.5
            if overlap > best_score:
                best_score = overlap
                best_sent = sent
                best_idx = i

        if best_score <= 0:
            return {
                "ok": True,
                "result": "No answer found",
                "answer": "No answer found",
                "confidence": 0.0,
                "span_start": -1,
                "span_end": -1
            }

        # Find the most relevant span within the best sentence
        words = best_sent.split()
        best_span_score = -1
        best_span = best_sent
        span_start_char = context.find(best_sent)
        span_end_char = span_start_char + len(best_sent)

        # Try to narrow down to a shorter span
        for window in range(3, min(len(words) + 1, 15)):
            for start in range(len(words) - window + 1):
                span = " ".join(words[start:start + window])
                span_tokens = set(_tokenize(span)) - STOPWORDS
                score = len(q_tokens & span_tokens)
                if score > best_span_score:
                    best_span_score = score
                    best_span = span
                    span_start_char = context.find(span)
                    span_end_char = span_start_char + len(span)

        confidence = round(min(1.0, best_score / max(len(q_tokens), 1)), 4)
        return {
            "ok": True,
            "result": best_span,
            "answer": best_span,
            "confidence": confidence,
            "span_start": span_start_char,
            "span_end": span_end_char,
            "source_sentence": best_sent
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
