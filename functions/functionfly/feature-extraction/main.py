import re
import math
from collections import Counter

STOPWORDS = {"a","an","the","is","are","was","were","be","been","have","has","had","do","does","did","will","would","could","should","to","of","in","for","on","with","at","by","from","and","but","or","not","this","that","it","its","i","you","he","she","we","they"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        feature_types = event.get("feature_types", ["statistical", "lexical", "syntactic"])
        features = {}
        words = re.findall(r'\b\w+\b', text)
        sentences = re.split(r'(?<=[.!?])\s+', text.strip())
        chars = list(text)

        if "statistical" in feature_types:
            word_lengths = [len(w) for w in words]
            features["statistical"] = {
                "char_count": len(text),
                "word_count": len(words),
                "sentence_count": len(sentences),
                "avg_word_length": round(sum(word_lengths) / len(word_lengths), 4) if word_lengths else 0,
                "avg_sentence_length": round(len(words) / len(sentences), 4) if sentences else 0,
                "unique_word_ratio": round(len(set(w.lower() for w in words)) / len(words), 4) if words else 0,
                "punctuation_ratio": round(sum(1 for c in text if c in '.,!?;:') / len(text), 4) if text else 0,
                "digit_ratio": round(sum(1 for c in text if c.isdigit()) / len(text), 4) if text else 0,
                "uppercase_ratio": round(sum(1 for c in text if c.isupper()) / len(text), 4) if text else 0,
            }

        if "lexical" in feature_types:
            lower_words = [w.lower() for w in words]
            content_words = [w for w in lower_words if w not in STOPWORDS]
            freq = Counter(lower_words)
            features["lexical"] = {
                "vocabulary_size": len(set(lower_words)),
                "type_token_ratio": round(len(set(lower_words)) / len(lower_words), 4) if lower_words else 0,
                "content_word_ratio": round(len(content_words) / len(lower_words), 4) if lower_words else 0,
                "stopword_ratio": round((len(lower_words) - len(content_words)) / len(lower_words), 4) if lower_words else 0,
                "hapax_legomena": sum(1 for w, c in freq.items() if c == 1),
                "top_words": [{"word": w, "count": c} for w, c in freq.most_common(10)],
            }

        if "syntactic" in feature_types:
            features["syntactic"] = {
                "question_count": text.count('?'),
                "exclamation_count": text.count('!'),
                "comma_count": text.count(','),
                "semicolon_count": text.count(';'),
                "colon_count": text.count(':'),
                "parentheses_count": text.count('(') + text.count(')'),
                "quote_count": text.count('"') + text.count("'"),
                "capitalized_words": sum(1 for w in words if w[0].isupper()),
                "all_caps_words": sum(1 for w in words if w.isupper() and len(w) > 1),
                "numbers": re.findall(r'\b\d+(?:\.\d+)?\b', text),
            }

        return {
            "ok": True,
            "result": features,
            "features": features,
            "feature_types": feature_types
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
