import hashlib

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    num_speakers = event.get("num_speakers")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        detected_speakers = num_speakers if num_speakers else (h[0] % 3) + 2
        segments = []
        t = 0.0
        for i in range(8):
            speaker_id = f"SPEAKER_{(i % detected_speakers) + 1:02d}"
            duration = 5.0 + (h[i % 32] / 255.0) * 15.0
            segments.append({
                "speaker": speaker_id,
                "start": round(t, 2),
                "end": round(t + duration, 2),
                "duration": round(duration, 2),
                "confidence": round(0.8 + (h[(i+1) % 32] / 255.0) * 0.2, 4)
            })
            t += duration + 0.5
        speakers = {}
        for seg in segments:
            spk = seg["speaker"]
            if spk not in speakers:
                speakers[spk] = {"id": spk, "total_duration": 0, "segment_count": 0}
            speakers[spk]["total_duration"] += seg["duration"]
            speakers[spk]["segment_count"] += 1
        return {
            "ok": True,
            "result": segments,
            "segments": segments,
            "speakers": list(speakers.values()),
            "num_speakers": detected_speakers,
            "total_duration": round(t, 2),
            "model": "mock-pyannote-v1",
            "note": "Mock diarization — for production use, integrate pyannote.audio or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
