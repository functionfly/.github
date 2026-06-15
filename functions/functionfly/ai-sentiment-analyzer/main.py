"""AI Sentiment Analyzer - Analyze text sentiment using keyword scoring."""
import re
from collections import Counter


POSITIVE_WORDS = {
    "good": 0.5, "great": 0.7, "excellent": 0.9, "amazing": 0.9, "wonderful": 0.9,
    "fantastic": 0.9, "awesome": 0.8, "love": 0.7, "loved": 0.7, "like": 0.3,
    "happy": 0.7, "pleased": 0.6, "delighted": 0.8, "satisfied": 0.6, "recommend": 0.6,
    "best": 0.8, "better": 0.5, "perfect": 0.9, "beautiful": 0.7, "nice": 0.4,
    "helpful": 0.5, "friendly": 0.5, "professional": 0.5, "quality": 0.5, "reliable": 0.5,
    "outstanding": 0.9, "superb": 0.9, "brilliant": 0.8, "impressive": 0.7, "exceptional": 0.9,
    "positive": 0.6, "success": 0.6, "successful": 0.6, "win": 0.6, "winning": 0.6,
    "excited": 0.7, "exciting": 0.7, "thrilled": 0.8, "glad": 0.6, "joy": 0.7,
    "enjoy": 0.5, "enjoyed": 0.5, "fun": 0.5, "easy": 0.4, "simple": 0.3,
    "fast": 0.4, "quick": 0.4, "efficient": 0.5, "effective": 0.5, "powerful": 0.6,
    "thank": 0.4, "thanks": 0.4, "appreciate": 0.5, "grateful": 0.6, "blessed": 0.6,
    "cool": 0.4, "fresh": 0.3, "modern": 0.3, "innovative": 0.5, "creative": 0.5,
    "intuitive": 0.5, "seamless": 0.6, "smooth": 0.5, "robust": 0.5, "strong": 0.5,
    "useful": 0.5, "valuable": 0.5, "beneficial": 0.5, "valuable": 0.5, "worthwhile": 0.5
}

NEGATIVE_WORDS = {
    "bad": -0.5, "terrible": -0.9, "awful": -0.9, "horrible": -0.9, "worst": -0.9,
    "hate": -0.8, "hated": -0.8, "dislike": -0.5, "disappointed": -0.7, "disappointing": -0.7,
    "sad": -0.6, "unhappy": -0.7, "frustrated": -0.7, "frustrating": -0.7, "angry": -0.8,
    "poor": -0.5, "mediocre": -0.5, "useless": -0.8, "worthless": -0.8, "useless": -0.8,
    "slow": -0.4, "broken": -0.7, "fail": -0.6, "failed": -0.6, "failure": -0.7,
    "problem": -0.5, "problems": -0.5, "issue": -0.4, "issues": -0.4, "bug": -0.5,
    "bugs": -0.5, "error": -0.5, "errors": -0.5, "mistake": -0.5, "mistakes": -0.5,
    "wrong": -0.5, "incorrect": -0.5, "fault": -0.5, "faulty": -0.6, "defective": -0.7,
    "annoying": -0.6, "annoyed": -0.6, "irritating": -0.6, "irritated": -0.6, "difficult": -0.4,
    "hard": -0.3, "complicated": -0.4, "confusing": -0.5, "confused": -0.5, "complicated": -0.4,
    "expensive": -0.4, "overpriced": -0.6, "cheap": -0.3, "unreliable": -0.6, "unstable": -0.5,
    "crash": -0.6, "crashes": -0.6, "crashed": -0.6, "laggy": -0.5, "lag": -0.4,
    "sucks": -0.7, "sucked": -0.7, "pathetic": -0.7, "disgusting": -0.8, "gross": -0.6,
    "never": -0.3, "regret": -0.5, "regretful": -0.5, "waste": -0.5, "wasted": -0.5,
    "scam": -0.8, "fraud": -0.8, "fake": -0.6, "scam": -0.8, "ripoff": -0.7,
    "avoid": -0.5, "skip": -0.3, "skipping": -0.3, "trash": -0.7, "garbage": -0.7
}


def handler(event):
    try:
        text = event.get("text", "")

        if not text:
            return {"ok": False, "error": "text is required"}

        clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
        words = clean_text.split()

        positive_found = []
        negative_found = []
        total_score = 0

        for word in words:
            if word in POSITIVE_WORDS:
                positive_found.append(word)
                total_score += POSITIVE_WORDS[word]
            elif word in NEGATIVE_WORDS:
                negative_found.append(word)
                total_score += NEGATIVE_WORDS[word]

        word_count = len(words)
        sentiment_count = len(positive_found) + len(negative_found)

        if sentiment_count == 0:
            sentiment = "neutral"
            score = 0.0
            confidence = 0.0
        else:
            avg_score = total_score / sentiment_count
            score = round(avg_score, 2)

            if score > 0.2:
                sentiment = "positive"
            elif score < -0.2:
                sentiment = "negative"
            else:
                sentiment = "neutral"

            intensity = min(abs(avg_score), 1.0)
            confidence = round(0.5 + (intensity * 0.5) * min(sentiment_count / 10, 1), 2)
            confidence = min(confidence, 1.0)

        return {
            "ok": True,
            "sentiment": sentiment,
            "score": score,
            "confidence": confidence,
            "word_count": word_count,
            "positive_words": positive_found,
            "negative_words": negative_found
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
