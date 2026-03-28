import hashlib

ACTIONS = [
    "walking", "running", "jumping", "sitting", "standing", "lying down",
    "eating", "drinking", "talking", "laughing", "crying", "waving",
    "clapping", "dancing", "swimming", "cycling", "driving", "cooking",
    "reading", "writing", "typing", "playing guitar", "playing piano",
    "throwing", "catching", "kicking", "punching", "hugging", "shaking hands",
    "climbing", "falling", "pushing", "pulling", "lifting", "carrying",
    "opening door", "closing door", "answering phone", "taking photo",
    "exercising", "stretching", "yoga", "martial arts", "playing sports"
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
        for i in range(min(top_k, len(ACTIONS))):
            idx = (h[i % 32] + i * 11) % len(ACTIONS)
            while idx in used:
                idx = (idx + 1) % len(ACTIONS)
            used.add(idx)
            score = round(max(0.01, (h[i % 32] / 255.0) * (1.0 - i * 0.15)), 4)
            predictions.append({"action": ACTIONS[idx], "score": score, "rank": i + 1})
        predictions.sort(key=lambda x: x["score"], reverse=True)
        return {
            "ok": True,
            "result": predictions,
            "predictions": predictions,
            "top_action": predictions[0]["action"] if predictions else None,
            "model": "mock-action-recognition-v1",
            "note": "Mock action recognition — for production use, integrate SlowFast or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
