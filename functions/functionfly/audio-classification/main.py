import hashlib

AUDIO_CLASSES = [
    "speech", "music", "noise", "silence", "laughter", "applause", "crying",
    "dog_bark", "cat_meow", "bird_chirp", "rain", "thunder", "wind", "ocean_waves",
    "traffic", "car_horn", "siren", "alarm", "phone_ring", "keyboard_typing",
    "door_knock", "footsteps", "crowd", "explosion", "gunshot", "glass_breaking",
    "water_dripping", "fire_crackling", "engine", "helicopter", "airplane",
    "classical_music", "rock_music", "jazz", "pop_music", "electronic_music",
]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    top_k = int(event.get("top_k", 5))
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        predictions = []
        used = set()
        for i in range(min(top_k, len(AUDIO_CLASSES))):
            idx = (h[i % 32] + i * 7) % len(AUDIO_CLASSES)
            while idx in used:
                idx = (idx + 1) % len(AUDIO_CLASSES)
            used.add(idx)
            score = round(max(0.01, (h[i % 32] / 255.0) * (1.0 - i * 0.15)), 4)
            predictions.append({"label": AUDIO_CLASSES[idx], "score": score, "rank": i + 1})
        predictions.sort(key=lambda x: x["score"], reverse=True)
        return {
            "ok": True,
            "result": predictions,
            "predictions": predictions,
            "top_label": predictions[0]["label"] if predictions else None,
            "model": "mock-audio-classifier-v1",
            "note": "Mock audio classification — for production use, integrate PANNs or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
