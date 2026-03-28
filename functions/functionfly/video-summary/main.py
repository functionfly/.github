import hashlib

SUMMARY_TEMPLATES = [
    "The video shows {action} in a {setting} environment. The main subject appears to be {subject}.",
    "This {duration}-second clip features {subject} {action}. The scene takes place in {setting}.",
    "A {setting} video depicting {subject} engaged in {action}. The content appears to be {category}.",
    "The footage captures {action} with {subject} as the primary focus in a {setting} setting.",
]

SUBJECTS = ["a person", "multiple people", "an animal", "a vehicle", "an object", "a group"]
ACTIONS = ["performing an activity", "moving through the scene", "interacting with the environment", "demonstrating a skill", "engaging in conversation"]
SETTINGS = ["indoor", "outdoor", "urban", "natural", "studio", "public"]
CATEGORIES = ["educational", "entertainment", "documentary", "tutorial", "news", "personal"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    video_url = event.get("video_url", "")
    num_keyframes = int(event.get("num_keyframes", 5))
    if not video_url:
        return {"ok": False, "error": "video_url is required"}
    try:
        h = hashlib.sha256(video_url.encode()).digest()
        template = SUMMARY_TEMPLATES[h[0] % len(SUMMARY_TEMPLATES)]
        duration = 30 + (h[1] / 255.0) * 270
        summary = template.format(
            action=ACTIONS[h[2] % len(ACTIONS)],
            setting=SETTINGS[h[3] % len(SETTINGS)],
            subject=SUBJECTS[h[4] % len(SUBJECTS)],
            duration=round(duration),
            category=CATEGORIES[h[5] % len(CATEGORIES)]
        )
        keyframes = [{"timestamp": round(duration * i / num_keyframes, 2), "description": f"Scene {i+1}", "thumbnail_url": f"mock://keyframe/{i}"} for i in range(num_keyframes)]
        return {
            "ok": True,
            "result": summary,
            "summary": summary,
            "keyframes": keyframes,
            "duration_seconds": round(duration, 2),
            "model": "mock-video-summary-v1",
            "note": "Mock video summarization — for production use, integrate a video understanding model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
