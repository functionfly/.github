POSITIVE_WORDS = {
    "excellent", "great", "amazing", "fantastic", "wonderful", "perfect", "love", "loved",
    "awesome", "best", "outstanding", "superb", "brilliant", "incredible", "recommend",
    "happy", "pleased", "satisfied", "good", "nice", "quality", "fast", "easy", "helpful",
    "beautiful", "comfortable", "durable", "reliable", "affordable", "value", "worth"
}
NEGATIVE_WORDS = {
    "terrible", "horrible", "awful", "bad", "poor", "worst", "disappointed", "disappointing",
    "broken", "defective", "slow", "expensive", "overpriced", "useless", "waste", "regret",
    "return", "refund", "broken", "damaged", "fake", "cheap", "nasty", "hate", "dislike",
    "frustrating", "difficult", "complicated", "unreliable", "misleading", "scam"
}
NEGATION_WORDS = {"not", "no", "never", "nothing", "neither", "nor", "hardly", "barely", "doesn't", "don't", "isn't", "wasn't"}


def handler(event):
    review = event.get("review") if isinstance(event, dict) else None
    if not review:
        return {"ok": False, "error": "review is required"}
    try:
        text = str(review).lower()
        words = text.split()
        pos_count = 0
        neg_count = 0
        found_positive = []
        found_negative = []
        for i, word in enumerate(words):
            clean = word.strip(".,!?\"'()[]")
            negated = i > 0 and words[i-1].strip(".,!?\"'") in NEGATION_WORDS
            if clean in POSITIVE_WORDS:
                if negated:
                    neg_count += 1
                    found_negative.append(f"not {clean}")
                else:
                    pos_count += 1
                    found_positive.append(clean)
            elif clean in NEGATIVE_WORDS:
                if negated:
                    pos_count += 1
                    found_positive.append(f"not {clean}")
                else:
                    neg_count += 1
                    found_negative.append(clean)
        total = pos_count + neg_count
        score = round((pos_count - neg_count) / max(total, 1), 4)
        if score > 0.25:
            sentiment = "positive"
        elif score < -0.25:
            sentiment = "negative"
        else:
            sentiment = "neutral"
        return {
            "ok": True,
            "result": sentiment,
            "sentiment": sentiment,
            "score": score,
            "positive_count": pos_count,
            "negative_count": neg_count,
            "positive_words": found_positive[:10],
            "negative_words": found_negative[:10]
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
