"""AI Text Summarizer - Extractive summarization using word frequency scoring."""
import re
from collections import Counter


def score_sentence(sentence, word_freq, total_words):
    words = re.sub(r'[^\w\s]', ' ', sentence.lower()).split()
    if not words:
        return 0
    score = sum(word_freq.get(w, 0) for w in words) / len(words)
    position_bonus = 1.0
    if len(words) > 0:
        position_bonus = 1.0 + (0.1 * max(0, 5 - abs(2 - words[0])))
    return score * position_bonus


def extractive_summarize(text, max_sentences, style):
    sentences = re.split(r'(?<=[.!?])\s+', text)
    sentences = [s.strip() for s in sentences if s.strip()]

    if len(sentences) <= max_sentences:
        return sentences, sentences

    clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
    words = [w for w in clean_text.split() if len(w) > 3]
    word_freq = Counter(words)
    total_words = len(words)

    if total_words == 0:
        return sentences[:max_sentences], sentences

    for word in word_freq:
        word_freq[word] = word_freq[word] / total_words

    scored_sentences = []
    for i, sentence in enumerate(sentences):
        score = score_sentence(sentence, word_freq, total_words)
        scored_sentences.append((i, sentence, score))

    scored_sentences.sort(key=lambda x: x[2], reverse=True)
    top_indices = sorted([idx for idx, _, _ in scored_sentences[:max_sentences]])

    top_sentences = [sentences[i] for i in top_indices]

    if style == "bullet":
        summary = "\n".join(f"• {s}" for s in top_sentences)
    else:
        summary = " ".join(top_sentences)

    return summary, top_sentences


def handler(event):
    try:
        text = event.get("text", "")
        max_sentences = event.get("max_sentences", 3)
        style = event.get("style", "paragraph")

        if not text:
            return {"ok": False, "error": "text is required"}
        if not isinstance(max_sentences, int) or max_sentences < 1:
            return {"ok": False, "error": "max_sentences must be a positive integer"}
        if style not in ["bullet", "paragraph"]:
            return {"ok": False, "error": "style must be bullet or paragraph"}

        original_words = text.split()
        original_word_count = len(original_words)

        summary, key_sentences = extractive_summarize(text, max_sentences, style)

        summary_words = summary.split()
        summary_word_count = len(summary_words)

        compression_ratio = round(summary_word_count / max(original_word_count, 1), 2)

        return {
            "ok": True,
            "summary": summary,
            "original_word_count": original_word_count,
            "summary_word_count": summary_word_count,
            "compression_ratio": compression_ratio,
            "key_sentences": key_sentences,
            "style": style,
            "max_sentences": max_sentences
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
