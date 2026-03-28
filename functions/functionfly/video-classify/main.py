import hashlib

VIDEO_CLASSES = [
    "sports", "music", "news", "documentary", "comedy", "drama", "action", "animation",
    "tutorial", "vlog", "gaming", "cooking", "travel", "fitness", "science", "technology",
    "nature", "animals", "art", "fashion", "politics", "education", "kids", "horror",
    "romance", "thriller", "sci-fi", "fantasy", "biography", "history"
]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    video_url = event.get("video_url", "")
    top_k = int(event.get("top_k", 5))
    if not video_url:
        return {"ok": False, "error": "video_url is required"}
    try:
        h = hashlib.sha256(video_url.encode()).digest()
        predictions = []
        used = set()
        for i in range(min(top_k, len(VIDEO_CLASSES))):
            idx = (h[i % 32] + i * 7) % len(VIDEO_CLASSES)
            while idx in used:
                idx = (idx + 1) % len(VIDEO_CLASSES)
            used.add(idx)
            score = round(max(0.01, (h[i % 32] / 255.0) * (1.0 - i * 0.15)), 4)
            predictions.append({"label": VIDEO_CLASSES[idx], "score": score, "rank": i + 1})
        predictions.sort(key=lambda x: x["score"], reverse=True)
        return {
            "ok": True,
            "result": predictions,
            "predictions": predictions,
            "top_label": predictions[0]["label"] if predictions else None,
            "model": "mock-video-classifier-v1",
            "note": "Mock video classification — for production use, integrate VideoMAE or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
