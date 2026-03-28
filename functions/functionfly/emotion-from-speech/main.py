import hashlib

SPEECH_EMOTIONS = ["neutral", "happy", "sad", "angry", "fearful", "disgusted", "surprised", "calm", "excited", "bored"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        scores = {}
        total = 0
        for i, emotion in enumerate(SPEECH_EMOTIONS):
            score = h[i % 32] / 255.0
            scores[emotion] = score
            total += score
        normalized = {e: round(s / total, 4) for e, s in scores.items()}
        dominant = max(normalized, key=normalized.get)
        acoustic_features = {
            "pitch_mean": round(100 + (h[10] / 255.0) * 200, 2),
            "pitch_std": round(10 + (h[11] / 255.0) * 50, 2),
            "energy_mean": round(0.1 + (h[12] / 255.0) * 0.9, 4),
            "speech_rate": round(2 + (h[13] / 255.0) * 4, 2),
            "pause_ratio": round(0.1 + (h[14] / 255.0) * 0.4, 4),
        }
        return {
            "ok": True,
            "result": {"dominant": dominant, "scores": normalized},
            "emotions": normalized,
            "dominant": dominant,
            "confidence": normalized[dominant],
            "acoustic_features": acoustic_features,
            "model": "mock-ser-v1",
            "note": "Mock speech emotion recognition — for production use, integrate SER model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
