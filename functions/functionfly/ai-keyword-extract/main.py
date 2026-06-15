"""AI Keyword Extractor - Extract keywords and phrases from text."""
import re
from collections import Counter


STOP_WORDS = {
    "a", "an", "the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with",
    "by", "from", "as", "is", "was", "are", "were", "been", "be", "have", "has", "had",
    "do", "does", "did", "will", "would", "could", "should", "may", "might", "must",
    "shall", "can", "need", "dare", "ought", "used", "it", "its", "this", "that", "these",
    "those", "i", "you", "he", "she", "we", "they", "what", "which", "who", "whom",
    "whose", "where", "when", "why", "how", "all", "each", "every", "both", "few",
    "more", "most", "other", "some", "such", "no", "nor", "not", "only", "own", "same",
    "so", "than", "too", "very", "just", "also", "now", "here", "there", "then"
}


def calculate_word_scores(words, max_keywords):
    word_freq = Counter(words)
    total_words = len(words)
    unique_words = len(word_freq)

    if total_words == 0:
        return []

    scored_words = []
    for word, freq in word_freq.items():
        if len(word) < 2:
            continue
        score = (freq / total_words) * (1 + (freq / max(1, unique_words)))
        scored_words.append((word, round(score, 4)))

    scored_words.sort(key=lambda x: x[1], reverse=True)
    return scored_words[:max_keywords]


def extract_phrases(text, words):
    phrases = []
    word_list = words
    for phrase_len in [2, 3]:
        for i in range(len(word_list) - phrase_len + 1):
            phrase = " ".join(word_list[i:i + phrase_len])
            if len(phrase) > 6:
                phrases.append(phrase)
    phrase_freq = Counter(phrases)
    return [phrase for phrase, count in phrase_freq.most_common(10)] if phrase_freq else []


def handler(event):
    try:
        text = event.get("text", "")
        max_keywords = event.get("max_keywords", 10)
        include_phrases = event.get("include_phrases", True)

        if not text:
            return {"ok": False, "error": "text is required"}
        if not isinstance(max_keywords, int) or max_keywords < 1:
            return {"ok": False, "error": "max_keywords must be a positive integer"}

        clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
        words = [w for w in clean_text.split() if w not in STOP_WORDS and len(w) > 2]

        if not words:
            return {
                "ok": True,
                "keywords": [],
                "key_phrases": [] if include_phrases else None,
                "word_count": len(text.split()),
                "unique_words": len(set(text.lower().split()))
            }

        scored_words = calculate_word_scores(words, max_keywords)
        keywords = [{"word": word, "score": score} for word, score in scored_words]

        key_phrases = None
        if include_phrases:
            phrases = extract_phrases(text, words)
            key_phrases = phrases[:10]

        return {
            "ok": True,
            "keywords": keywords,
            "key_phrases": key_phrases,
            "word_count": len(text.split()),
            "unique_words": len(set(text.lower().split()))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
